package group

import (
	"context"
	"fmt"
	"im-system/gateway/connection"
	"im-system/pkg/logger"
	"im-system/pkg/redis"
	"im-system/proto"
	"strconv"

	"go.uber.org/zap"
)

const groupMembersKeyPrefix = "im:group:members:" // im:group:members:{groupID} -> Set of uid

func groupMembersKey(groupID int64) string {
	return fmt.Sprintf("%s%d", groupMembersKeyPrefix, groupID)
}

type Service struct {
	manager *connection.Manager
}

func NewService(manager *connection.Manager) *Service {
	return &Service{manager: manager}
}

// Deliver 群聊写扩散：向群内所有成员投递消息
// 在线的直接推送，离线的存入各自的离线队列
func (s *Service) Deliver(msg *proto.Message, offline func(uid int64, msg *proto.Message)) {
	members, err := s.getMembers(msg.ToID)
	if err != nil {
		logger.Error("get group members failed", zap.Int64("group_id", msg.ToID), zap.Error(err))
		return
	}

	data, _ := proto.Encode(msg)
	for _, uid := range members {
		if uid == msg.FromUID {
			continue // 不给自己推
		}
		if !s.manager.SendToUser(uid, data) {
			if offline != nil {
				offline(uid, msg)
			}
		}
	}
}

// AddMember 加入群组
func (s *Service) AddMember(groupID, uid int64) error {
	ctx := context.Background()
	return redis.Client.SAdd(ctx, groupMembersKey(groupID), uid).Err()
}

// RemoveMember 退出群组
func (s *Service) RemoveMember(groupID, uid int64) error {
	ctx := context.Background()
	return redis.Client.SRem(ctx, groupMembersKey(groupID), uid).Err()
}

// getMembers 从 Redis 获取群成员列表
func (s *Service) getMembers(groupID int64) ([]int64, error) {
	ctx := context.Background()
	vals, err := redis.Client.SMembers(ctx, groupMembersKey(groupID)).Result()
	if err != nil {
		return nil, err
	}
	uids := make([]int64, 0, len(vals))
	for _, v := range vals {
		uid, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			uids = append(uids, uid)
		}
	}
	return uids, nil
}
