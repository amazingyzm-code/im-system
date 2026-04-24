package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"im-system/proto"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

var (
	addr        = flag.String("addr", "localhost:8080", "server address")
	connections = flag.Int("c", 100, "concurrent connections")
	duration    = flag.Duration("d", 30*time.Second, "test duration")
	msgInterval = flag.Duration("i", 200*time.Millisecond, "message interval per connection")
)

var (
	totalSent    int64
	totalRecv    int64
	totalErrors  int64
	connErrors   int64 // 连接阶段错误
)

func main() {
	flag.Parse()

	fmt.Printf("benchmark: %d connections, %s duration\n", *connections, *duration)
	fmt.Println("registering test users...")

	tokens := make([]string, *connections)
	for i := 0; i < *connections; i++ {
		username := fmt.Sprintf("bench_user_%d", i)
		token, err := loginOrRegister(*addr, username, "bench123")
		if err != nil {
			fmt.Printf("auth failed for %s: %v\n", username, err)
			return
		}
		tokens[i] = token
	}
	fmt.Println("all users ready, starting benchmark...")

	var wg sync.WaitGroup
	stop := make(chan struct{})

	for i := 0; i < *connections; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			runClient(tokens[idx], idx, stop)
		}(i)
	}

	// 定时打印进度
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		start := time.Now()
		for range ticker.C {
			elapsed := time.Since(start).Seconds()
			sent := atomic.LoadInt64(&totalSent)
			recv := atomic.LoadInt64(&totalRecv)
			errs := atomic.LoadInt64(&totalErrors)
			fmt.Printf("[%.0fs] sent=%d recv=%d errors=%d send_qps=%.0f\n",
				elapsed, sent, recv, errs, float64(sent)/elapsed)
		}
	}()

	time.Sleep(*duration)
	close(stop)
	ticker.Stop()
	wg.Wait()

	elapsed := duration.Seconds()
	sent := atomic.LoadInt64(&totalSent)
	recv := atomic.LoadInt64(&totalRecv)
	errs := atomic.LoadInt64(&totalErrors)

	fmt.Println("\n========== RESULT ==========")
	fmt.Printf("duration:      %s\n", *duration)
	fmt.Printf("connections:   %d\n", *connections)
	fmt.Printf("total sent:    %d\n", sent)
	fmt.Printf("total recv:    %d\n", recv)
	fmt.Printf("msg errors:    %d\n", errs)
	fmt.Printf("conn errors:   %d\n", atomic.LoadInt64(&connErrors))
	fmt.Printf("send QPS:      %.0f\n", float64(sent)/elapsed)
	fmt.Printf("recv QPS:      %.0f\n", float64(recv)/elapsed)
}

func runClient(token string, idx int, stop chan struct{}) {
	conn, _, err := websocket.DefaultDialer.Dial("ws://"+*addr+"/ws", nil)
	if err != nil {
		atomic.AddInt64(&connErrors, 1)
		return
	}
	defer conn.Close()

	authMsg := proto.AuthMessage{MsgType: proto.MsgTypeAuth, Token: token}
	data, _ := json.Marshal(authMsg)
	conn.WriteMessage(websocket.TextMessage, data)

	// 接收 goroutine
	go func() {
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				return
			}
			atomic.AddInt64(&totalRecv, 1)
		}
	}()

	// 发送消息：发给下一个用户（形成环形）
	toUID := int64((idx+1)%(*connections)) + 1
	ticker := time.NewTicker(*msgInterval)
	defer ticker.Stop()

	var seq int64
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			seq++
			msg := proto.Message{
				MsgType:    proto.MsgTypeText,
				TargetType: proto.TargetTypeSingle,
				ToID:       toUID,
				Content:    "bench",
				Timestamp:  time.Now().UnixMilli(),
				Seq:        seq,
			}
			data, _ := json.Marshal(msg)
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				atomic.AddInt64(&totalErrors, 1)
				return
			}
			atomic.AddInt64(&totalSent, 1)
		}
	}
}

func loginOrRegister(addr, username, password string) (string, error) {
	body, _ := json.Marshal(map[string]string{"username": username, "password": password})
	token, err := postJSON("http://"+addr+"/api/user/login", body)
	if err == nil {
		return token, nil
	}
	postJSON("http://"+addr+"/api/user/register", body)
	return postJSON("http://"+addr+"/api/user/login", body)
}

func postJSON(url string, body []byte) (string, error) {
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}
	var result map[string]string
	json.Unmarshal(data, &result)
	return result["token"], nil
}
