package connection

import (
	"im-system/pkg/logger"
	"sync"

	"go.uber.org/zap"
)

// Manager 管理所有在线 WebSocket 连接
type Manager struct {
	mu    sync.RWMutex
	conns map[int64]*Connection // uid -> Connection
}

func NewManager() *Manager {
	return &Manager{
		conns: make(map[int64]*Connection),
	}
}

func (m *Manager) Register(conn *Connection) {
	m.mu.Lock()
	// 同一用户重复登录，踢掉旧连接
	if old, ok := m.conns[conn.UID]; ok {
		old.Close()
		logger.Info("kickout old connection", zap.Int64("uid", conn.UID))
	}
	m.conns[conn.UID] = conn
	m.mu.Unlock()
	logger.Info("user online", zap.Int64("uid", conn.UID), zap.Int("online", m.OnlineCount()))
}

func (m *Manager) Unregister(uid int64) {
	m.mu.Lock()
	delete(m.conns, uid)
	m.mu.Unlock()
	logger.Info("user offline", zap.Int64("uid", uid), zap.Int("online", m.OnlineCount()))
}

// Get 获取某个用户的连接
func (m *Manager) Get(uid int64) (*Connection, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	conn, ok := m.conns[uid]
	return conn, ok
}

// SendToUser 向指定用户发送消息
func (m *Manager) SendToUser(uid int64, data []byte) bool {
	conn, ok := m.Get(uid)
	if !ok {
		return false
	}
	return conn.Send(data)
}

// Broadcast 向所有在线用户广播
func (m *Manager) Broadcast(data []byte) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, conn := range m.conns {
		conn.Send(data)
	}
}

func (m *Manager) OnlineCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.conns)
}

func (m *Manager) IsOnline(uid int64) bool {
	_, ok := m.Get(uid)
	return ok
}
