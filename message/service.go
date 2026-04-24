package message

import (
	"context"
	"im-system/gateway/connection"
	"im-system/group"
	"im-system/pkg/db"
	"im-system/pkg/logger"
	"im-system/pkg/mq"
	"im-system/pkg/snowflake"
	"im-system/proto"

	"go.uber.org/zap"
)

// OfflineStore 离线消息存储接口
type OfflineStore interface {
	Save(uid int64, msg *proto.Message) error
	Pull(uid int64, lastMsgID int64) ([]*proto.Message, error)
}

type Service struct {
	manager      *connection.Manager
	offline      OfflineStore
	groupService *group.Service
	idWorker     *snowflake.Worker
	producer     *mq.Producer // Kafka 生产者
}

func NewService(manager *connection.Manager, offline OfflineStore, groupSvc *group.Service, producer *mq.Producer) *Service {
	worker, _ := snowflake.NewWorker(1)
	return &Service{
		manager:      manager,
		offline:      offline,
		groupService: groupSvc,
		idWorker:     worker,
		producer:     producer,
	}
}

// Handle 收到客户端消息，分配 ID 后推入 Kafka
func (s *Service) Handle(conn *connection.Connection, msg *proto.Message) {
	msg.MsgID = s.idWorker.NextID()
	msg.FromUID = conn.UID

	// 先回 ACK，告知客户端消息已收到
	conn.SendAck(msg.Seq, msg.MsgID)

	// 异步持久化
	go func() {
		if err := db.SaveMessage(msg); err != nil {
			logger.Error("persist message failed", zap.Int64("msg_id", msg.MsgID), zap.Error(err))
		}
	}()

	// 推入 Kafka，由 Consumer 异步投递
	topic := mq.TopicSingleMsg
	if msg.TargetType == proto.TargetTypeGroup {
		topic = mq.TopicGroupMsg
	}

	if s.producer != nil {
		if err := s.producer.Publish(context.Background(), topic, msg); err != nil {
			logger.Error("kafka publish failed", zap.Error(err))
			// Kafka 失败降级：直接同步投递
			s.deliver(msg)
		}
	} else {
		s.deliver(msg)
	}
}

// Deliver 由 Kafka Consumer 调用，执行实际投递
func (s *Service) Deliver(msg *proto.Message) {
	s.deliver(msg)
}

func (s *Service) deliver(msg *proto.Message) {
	switch msg.TargetType {
	case proto.TargetTypeSingle:
		s.deliverSingle(msg)
	case proto.TargetTypeGroup:
		s.deliverGroup(msg)
	}
}

func (s *Service) deliverSingle(msg *proto.Message) {
	data, _ := proto.Encode(msg)
	if s.manager.SendToUser(msg.ToID, data) {
		logger.Info("single msg delivered",
			zap.Int64("from", msg.FromUID),
			zap.Int64("to", msg.ToID),
			zap.Int64("msg_id", msg.MsgID),
		)
		return
	}
	s.saveOffline(msg.ToID, msg)
}

func (s *Service) deliverGroup(msg *proto.Message) {
	if s.groupService == nil {
		return
	}
	s.groupService.Deliver(msg, func(uid int64, m *proto.Message) {
		s.saveOffline(uid, m)
	})
}

func (s *Service) saveOffline(uid int64, msg *proto.Message) {
	if s.offline == nil {
		return
	}
	if err := s.offline.Save(uid, msg); err != nil {
		logger.Error("save offline msg failed", zap.Int64("uid", uid), zap.Error(err))
		return
	}
	logger.Info("msg saved offline", zap.Int64("uid", uid), zap.Int64("msg_id", msg.MsgID))
}

// PullOffline 用户上线后拉取离线消息
func (s *Service) PullOffline(uid int64, lastMsgID int64) {
	if s.offline == nil {
		return
	}
	msgs, err := s.offline.Pull(uid, lastMsgID)
	if err != nil {
		logger.Error("pull offline failed", zap.Int64("uid", uid), zap.Error(err))
		return
	}
	for _, msg := range msgs {
		data, _ := proto.Encode(msg)
		s.manager.SendToUser(uid, data)
	}
	logger.Info("offline msgs pulled", zap.Int64("uid", uid), zap.Int("count", len(msgs)))
}
