package redis

import (
	"context"
	"fmt"
	"time"
)

const (
	onlineKeyPrefix = "im:online:"   // im:online:{uid} -> nodeID
	onlineTTL       = 2 * time.Minute
)

func onlineKey(uid int64) string {
	return fmt.Sprintf("%s%d", onlineKeyPrefix, uid)
}

// SetOnline 标记用户上线，记录所在节点
func SetOnline(ctx context.Context, uid int64, nodeID string) error {
	return Client.Set(ctx, onlineKey(uid), nodeID, onlineTTL).Err()
}

// SetOffline 标记用户下线
func SetOffline(ctx context.Context, uid int64) error {
	return Client.Del(ctx, onlineKey(uid)).Err()
}

// GetUserNode 获取用户所在节点，返回空字符串表示离线
func GetUserNode(ctx context.Context, uid int64) (string, error) {
	val, err := Client.Get(ctx, onlineKey(uid)).Result()
	if err != nil {
		return "", nil // 不在线
	}
	return val, nil
}

// RefreshOnline 刷新在线 TTL（心跳时调用）
func RefreshOnline(ctx context.Context, uid int64) error {
	return Client.Expire(ctx, onlineKey(uid), onlineTTL).Err()
}
