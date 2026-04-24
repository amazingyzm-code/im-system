package connection

import (
	"im-system/pkg/logger"
	"im-system/proto"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	heartbeatInterval = 30 * time.Second
	writeTimeout      = 10 * time.Second
	maxMessageSize    = 4096
)

// Connection 封装一个 WebSocket 连接
type Connection struct {
	UID      int64
	conn     *websocket.Conn
	send     chan []byte
	once     sync.Once
	closeCh  chan struct{}
}

func New(uid int64, conn *websocket.Conn) *Connection {
	return &Connection{
		UID:     uid,
		conn:    conn,
		send:    make(chan []byte, 2048),
		closeCh: make(chan struct{}),
	}
}

// Send 将消息放入发送队列（非阻塞）
func (c *Connection) Send(data []byte) bool {
	select {
	case c.send <- data:
		return true
	default:
		logger.Warn("connection send buffer full", zap.Int64("uid", c.UID))
		return false
	}
}

// Close 关闭连接（幂等）
func (c *Connection) Close() {
	c.once.Do(func() {
		close(c.closeCh)
		c.conn.Close()
	})
}

// ReadPump 持续读取客户端消息，收到消息后交给 handler 处理
func (c *Connection) ReadPump(handler func([]byte)) {
	defer c.Close()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(heartbeatInterval * 2))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(heartbeatInterval * 2))
		return nil
	})

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				logger.Warn("ws read error", zap.Int64("uid", c.UID), zap.Error(err))
			}
			return
		}
		handler(data)
	}
}

// WritePump 持续将 send 队列中的消息写给客户端，并定时发送心跳 Ping
func (c *Connection) WritePump() {
	ticker := time.NewTicker(heartbeatInterval)
	defer func() {
		ticker.Stop()
		c.Close()
	}()

	for {
		select {
		case data, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, nil)
				return
			}
			// 批量写：把队列里积压的消息一次性发出去
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(data)
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte("\n"))
				w.Write(<-c.send)
			}
			if err := w.Close(); err != nil {
				logger.Error("ws write error", zap.Int64("uid", c.UID), zap.Error(err))
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-c.closeCh:
			return
		}
	}
}

// SendAck 向客户端发送 ACK
func (c *Connection) SendAck(seq, msgID int64) {
	ack := proto.AckMessage{
		MsgType: proto.MsgTypeAck,
		Seq:     seq,
		MsgID:   msgID,
	}
	data, _ := proto.Encode(ack)
	c.Send(data)
}
