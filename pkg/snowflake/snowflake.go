package snowflake

import (
	"errors"
	"sync"
	"time"
)

// 雪花算法：64位 = 1位符号 + 41位时间戳 + 10位机器ID + 12位序列号
const (
	epoch          = int64(1700000000000) // 自定义起始时间戳（毫秒）
	workerBits     = 10
	sequenceBits   = 12
	maxWorkerID    = -1 ^ (-1 << workerBits)   // 1023
	maxSequence    = -1 ^ (-1 << sequenceBits)  // 4095
	timeShift      = workerBits + sequenceBits   // 22
	workerShift    = sequenceBits                // 12
)

type Worker struct {
	mu        sync.Mutex
	workerID  int64
	sequence  int64
	lastStamp int64
}

func NewWorker(workerID int64) (*Worker, error) {
	if workerID < 0 || workerID > maxWorkerID {
		return nil, errors.New("worker id out of range")
	}
	return &Worker{workerID: workerID}, nil
}

func (w *Worker) NextID() int64 {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now().UnixMilli()
	if now == w.lastStamp {
		w.sequence = (w.sequence + 1) & maxSequence
		if w.sequence == 0 {
			for now <= w.lastStamp {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		w.sequence = 0
	}
	w.lastStamp = now
	return (now-epoch)<<timeShift | w.workerID<<workerShift | w.sequence
}
