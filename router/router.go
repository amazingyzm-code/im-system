package router

import (
	"context"
	"encoding/json"
	"im-system/gateway/connection"
	"im-system/pkg/logger"
	"im-system/pkg/redis"
	"im-system/proto"

	"go.uber.org/zap"
)

const channelPrefix = "im:node:" // im:node:{nodeID}

// Router 负责跨节点消息路由
// 当目标用户不在本节点时，通过 Redis Pub/Sub 转发到对应节点
type Router struct {
	nodeID  string
	manager *connection.Manager
}

func New(nodeID string, manager *connection.Manager) *Router {
	return &Router{nodeID: nodeID, manager: manager}
}

// Publish 将消息发布到目标节点的频道
func (r *Router) Publish(targetNodeID string, msg *proto.Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return redis.Client.Publish(context.Background(), channelPrefix+targetNodeID, data).Err()
}

// Subscribe 订阅本节点频道，收到消息后投递给本地连接
func (r *Router) Subscribe(ctx context.Context) {
	channel := channelPrefix + r.nodeID
	sub := redis.Client.Subscribe(ctx, channel)
	defer sub.Close()

	logger.Info("router subscribed", zap.String("channel", channel))

	for {
		select {
		case <-ctx.Done():
			return
		case redisMsg, ok := <-sub.Channel():
			if !ok {
				return
			}
			var msg proto.Message
			if err := json.Unmarshal([]byte(redisMsg.Payload), &msg); err != nil {
				logger.Warn("router decode failed", zap.Error(err))
				continue
			}
			data, _ := proto.Encode(&msg)
			if !r.manager.SendToUser(msg.ToID, data) {
				logger.Warn("router deliver failed, user not on this node",
					zap.Int64("uid", msg.ToID))
			}
		}
	}
}
