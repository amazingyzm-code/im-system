package mq

import (
	"context"
	"encoding/json"
	"im-system/pkg/logger"
	"im-system/proto"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

const (
	TopicSingleMsg = "im.single.msg"
	TopicGroupMsg  = "im.group.msg"
)

// Producer 消息生产者
type Producer struct {
	writers map[string]*kafka.Writer
}

func NewProducer(brokers []string) *Producer {
	newWriter := func(topic string) *kafka.Writer {
		return &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.Hash{}, // 按 key hash 保证同一接收方顺序
			BatchSize:    100,           // 每批最多100条
			BatchTimeout: 2 * time.Millisecond,
			RequiredAcks: kafka.RequireOne,
			Async:        true, // 异步发送，不等 broker 确认，最大化吞吐
			Completion: func(messages []kafka.Message, err error) {
				if err != nil {
					logger.Warn("kafka async write failed", zap.Error(err))
				}
			},
		}
	}
	return &Producer{
		writers: map[string]*kafka.Writer{
			TopicSingleMsg: newWriter(TopicSingleMsg),
			TopicGroupMsg:  newWriter(TopicGroupMsg),
		},
	}
}

// Publish 发布消息到指定 topic
func (p *Producer) Publish(ctx context.Context, topic string, msg *proto.Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	w, ok := p.writers[topic]
	if !ok {
		return nil
	}
	return w.WriteMessages(ctx, kafka.Message{
		Key:   []byte{byte(msg.ToID), byte(msg.ToID >> 8)}, // 按接收方分区保证顺序
		Value: data,
	})
}

func (p *Producer) Close() {
	for _, w := range p.writers {
		w.Close()
	}
}

// Consumer 消息消费者，多 goroutine 并发消费
type Consumer struct {
	reader      *kafka.Reader
	handler     func(msg *proto.Message)
	workerCount int
}

func NewConsumer(brokers []string, topic, groupID string, handler func(msg *proto.Message)) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        brokers,
			Topic:          topic,
			GroupID:        groupID,
			MinBytes:       1 << 10,       // 1KB，攒够再拉
			MaxBytes:       1 << 20,       // 1MB
			MaxWait:        2 * time.Millisecond, // 最多等 2ms，保证低延迟
			CommitInterval: 500 * time.Millisecond,
			StartOffset:    kafka.LastOffset, // 只消费新消息
		}),
		handler:     handler,
		workerCount: 8, // 8个 worker 并发处理
	}
}

// Start 启动多 worker 并发消费
func (c *Consumer) Start(ctx context.Context) {
	msgCh := make(chan *proto.Message, 1024)

	// 启动 worker pool
	for i := 0; i < c.workerCount; i++ {
		go func() {
			for msg := range msgCh {
				c.handler(msg)
			}
		}()
	}

	// 单 goroutine 拉取，分发给 worker
	for {
		kafkaMsg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				close(msgCh)
				return
			}
			logger.Error("kafka read error", zap.Error(err))
			continue
		}

		var msg proto.Message
		if err := json.Unmarshal(kafkaMsg.Value, &msg); err != nil {
			logger.Warn("kafka decode error", zap.Error(err))
			continue
		}

		select {
		case msgCh <- &msg:
		default:
			// worker 全忙，直接同步处理防止消息丢失
			c.handler(&msg)
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
