package mq

import (
	"context"
	"net"
	"time"

	"github.com/segmentio/kafka-go"
)

func CreateTopics(brokers []string) error {
	conn, err := kafka.DialContext(context.Background(), "tcp", brokers[0])
	if err != nil {
		return err
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return err
	}

	controllerConn, err := kafka.DialContext(
		context.Background(), "tcp",
		net.JoinHostPort(controller.Host, string(rune(controller.Port+'0'))),
	)
	if err != nil {
		// 直接用原连接创建
		return createTopics(conn)
	}
	defer controllerConn.Close()
	return createTopics(controllerConn)
}

func createTopics(conn *kafka.Conn) error {
	topics := []kafka.TopicConfig{
		{Topic: TopicSingleMsg, NumPartitions: 8, ReplicationFactor: 1},
		{Topic: TopicGroupMsg, NumPartitions: 8, ReplicationFactor: 1},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = ctx
	return conn.CreateTopics(topics...)
}
