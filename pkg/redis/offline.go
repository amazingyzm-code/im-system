package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"im-system/proto"
	"time"
)

const (
	offlineKeyPrefix = "im:offline:" // im:offline:{uid} -> List of messages
	offlineMaxLen    = 200           // 每个用户最多缓存 200 条离线消息
	offlineTTL       = 7 * 24 * time.Hour
)

func offlineKey(uid int64) string {
	return fmt.Sprintf("%s%d", offlineKeyPrefix, uid)
}

// Save 保存离线消息
func (s *OfflineStore) Save(uid int64, msg *proto.Message) error {
	ctx := context.Background()
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	key := offlineKey(uid)
	pipe := Client.Pipeline()
	pipe.RPush(ctx, key, data)
	pipe.LTrim(ctx, key, -offlineMaxLen, -1) // 超出上限时丢弃最旧的
	pipe.Expire(ctx, key, offlineTTL)
	_, err = pipe.Exec(ctx)
	return err
}

// Pull 拉取离线消息（全量拉取后清空）
func (s *OfflineStore) Pull(uid int64, _ int64) ([]*proto.Message, error) {
	ctx := context.Background()
	key := offlineKey(uid)

	dataList, err := Client.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, err
	}
	Client.Del(ctx, key) // 拉取后清空

	msgs := make([]*proto.Message, 0, len(dataList))
	for _, d := range dataList {
		var msg proto.Message
		if err := json.Unmarshal([]byte(d), &msg); err == nil {
			msgs = append(msgs, &msg)
		}
	}
	return msgs, nil
}

// OfflineStore 实现 message.OfflineStore 接口
type OfflineStore struct{}
