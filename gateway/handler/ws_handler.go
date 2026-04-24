package handler

import (
	"im-system/gateway/connection"
	"im-system/pkg/logger"
	"im-system/proto"
	"im-system/user"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // 生产环境需校验 Origin
}

type Handler struct {
	manager    *connection.Manager
	msgHandler MsgHandler
	jwtSecret  string
	onOnline   func(uid int64)
}

// MsgHandler 是消息处理函数，由上层业务注入
type MsgHandler func(conn *connection.Connection, msg *proto.Message)

func New(manager *connection.Manager, msgHandler MsgHandler, jwtSecret string, onOnline func(uid int64)) *Handler {
	return &Handler{manager: manager, msgHandler: msgHandler, jwtSecret: jwtSecret, onOnline: onOnline}
}

// ServeWS 处理 WebSocket 连接请求
func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("ws upgrade failed", zap.Error(err))
		return
	}

	// 第一个包必须是鉴权包
	uid, err := h.auth(wsConn)
	if err != nil {
		logger.Warn("ws auth failed", zap.Error(err))
		wsConn.Close()
		return
	}

	conn := connection.New(uid, wsConn)
	h.manager.Register(conn)

	// 上线后触发离线消息拉取
	if h.onOnline != nil {
		go h.onOnline(uid)
	}

	go conn.WritePump()
	go func() {
		defer h.manager.Unregister(uid)
		conn.ReadPump(func(data []byte) {
			h.dispatch(conn, data)
		})
	}()
}

// auth 读取第一个包做鉴权，返回 uid
func (h *Handler) auth(wsConn *websocket.Conn) (int64, error) {
	wsConn.SetReadDeadline(time.Now().Add(10 * time.Second))
	_, data, err := wsConn.ReadMessage()
	if err != nil {
		return 0, err
	}
	wsConn.SetReadDeadline(time.Time{})

	var authMsg proto.AuthMessage
	if err := proto.Decode(data, &authMsg); err != nil {
		return 0, err
	}
	return user.ParseToken(authMsg.Token, h.jwtSecret)
}

// dispatch 根据消息类型分发处理
func (h *Handler) dispatch(conn *connection.Connection, data []byte) {
	var msg proto.Message
	if err := proto.Decode(data, &msg); err != nil {
		logger.Warn("decode message failed", zap.Int64("uid", conn.UID), zap.Error(err))
		return
	}

	switch msg.MsgType {
	case proto.MsgTypeHeartbeat:
		// 心跳直接忽略，Pong 由 WebSocket 层自动处理
	case proto.MsgTypeText, proto.MsgTypeImage:
		msg.FromUID = conn.UID
		msg.Timestamp = time.Now().UnixMilli()
		h.msgHandler(conn, &msg)
	default:
		logger.Warn("unknown msg type", zap.Int("type", msg.MsgType))
	}
}
