package proto

import "encoding/json"

// 消息类型
const (
	MsgTypeText      = 1  // 文本消息
	MsgTypeImage     = 2  // 图片消息
	MsgTypeAck       = 3  // ACK 确认
	MsgTypeHeartbeat = 4  // 心跳
	MsgTypeAuth      = 5  // 鉴权
	MsgTypeKickout   = 6  // 踢下线
)

// 消息目标类型
const (
	TargetTypeSingle = 1 // 单聊
	TargetTypeGroup  = 2 // 群聊
)

// Message 是客户端与服务端之间传输的基础消息结构
type Message struct {
	MsgID      int64  `json:"msg_id"`       // 消息唯一 ID（服务端生成）
	MsgType    int    `json:"msg_type"`     // 消息类型
	TargetType int    `json:"target_type"`  // 单聊/群聊
	FromUID    int64  `json:"from_uid"`     // 发送方 UID
	ToID       int64  `json:"to_id"`        // 接收方 UID 或群组 ID
	Content    string `json:"content"`      // 消息内容
	Timestamp  int64  `json:"timestamp"`    // 发送时间戳（毫秒）
	Seq        int64  `json:"seq"`          // 客户端序列号，用于 ACK
}

// AckMessage 是服务端回给客户端的 ACK
type AckMessage struct {
	MsgType int   `json:"msg_type"` // MsgTypeAck
	Seq     int64 `json:"seq"`      // 对应客户端的 seq
	MsgID   int64 `json:"msg_id"`   // 服务端分配的消息 ID
}

// AuthMessage 是客户端连接后发送的鉴权包
type AuthMessage struct {
	MsgType int    `json:"msg_type"` // MsgTypeAuth
	Token   string `json:"token"`
}

func Encode(v any) ([]byte, error) {
	return json.Marshal(v)
}

func Decode(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
