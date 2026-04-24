package main

import (
	"context"
	"fmt"
	"im-system/api"
	"im-system/gateway/connection"
	"im-system/gateway/handler"
	"im-system/group"
	"im-system/message"
	"im-system/pkg/config"
	"im-system/pkg/db"
	"im-system/pkg/limiter"
	"im-system/pkg/logger"
	"im-system/pkg/mq"
	rdb "im-system/pkg/redis"
	"im-system/proto"
	"im-system/router"
	"im-system/user"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

func main() {
	if err := config.Init("config.yaml"); err != nil {
		fmt.Printf("load config failed: %v\n", err)
		os.Exit(1)
	}

	logger.Init(config.Global.Server.Mode)
	defer logger.Sync()

	if err := rdb.Init(); err != nil {
		logger.Fatal("redis init failed", zap.Error(err))
	}
	logger.Info("redis connected")

	if err := db.Init(); err != nil {
		logger.Fatal("mysql init failed", zap.Error(err))
	}
	logger.Info("mysql connected")

	// Kafka
	brokers := config.Global.Kafka.Brokers
	if err := mq.CreateTopics(brokers); err != nil {
		logger.Warn("kafka create topics failed, will retry on use", zap.Error(err))
	}
	producer := mq.NewProducer(brokers)
	defer producer.Close()
	logger.Info("kafka producer ready")

	// 核心组件
	manager := connection.NewManager()
	groupSvc := group.NewService(manager)
	offlineStore := &rdb.OfflineStore{}
	msgService := message.NewService(manager, offlineStore, groupSvc, producer)

	// 启动 Kafka Consumer — 单聊
	singleConsumer := mq.NewConsumer(brokers, mq.TopicSingleMsg, "im-single-group", func(msg *proto.Message) {
		msgService.Deliver(msg)
	})
	// 启动 Kafka Consumer — 群聊
	groupConsumer := mq.NewConsumer(brokers, mq.TopicGroupMsg, "im-group-group", func(msg *proto.Message) {
		msgService.Deliver(msg)
	})

	ctx, cancel := context.WithCancel(context.Background())
	go singleConsumer.Start(ctx)
	go groupConsumer.Start(ctx)
	logger.Info("kafka consumers started")

	// 跨节点路由
	nodeID := config.Global.Server.NodeID
	r := router.New(nodeID, manager)
	go r.Subscribe(ctx)

	// 限流
	userLimiter := limiter.NewUserLimiterManager(10, 20)

	// WebSocket Handler
	wsHandler := handler.New(manager, func(conn *connection.Connection, msg *proto.Message) {
		if !userLimiter.Allow(conn.UID) {
			logger.Warn("rate limited", zap.Int64("uid", conn.UID))
			return
		}
		msgService.Handle(conn, msg)
	}, config.Global.JWT.Secret, func(uid int64) {
		msgService.PullOffline(uid, 0)
	})

	// HTTP 路由
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", wsHandler.ServeWS)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"status":"ok","online":%d}`, manager.OnlineCount())
	})
	userSvc := user.NewService(config.Global.JWT.Secret)
	user.NewHandler(userSvc).RegisterRoutes(mux)
	api.NewServer(groupSvc).RegisterRoutes(mux)

	addr := fmt.Sprintf(":%d", config.Global.Server.Port)
	logger.Info("im-system starting", zap.String("addr", addr), zap.String("node", nodeID))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := http.ListenAndServe(addr, mux); err != nil {
			logger.Fatal("server error", zap.Error(err))
		}
	}()

	<-quit
	cancel()
	singleConsumer.Close()
	groupConsumer.Close()
	logger.Info("im-system shutdown")
}
