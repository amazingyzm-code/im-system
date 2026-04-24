package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"im-system/proto"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var (
	addr     = flag.String("addr", "localhost:8080", "server address")
	username = flag.String("user", "", "username (will register+login automatically)")
	password = flag.String("pass", "123456", "password")
	token    = flag.String("token", "", "JWT token (skip login if provided)")
)

func main() {
	flag.Parse()

	jwtToken := *token
	if jwtToken == "" {
		if *username == "" {
			fmt.Println("provide -user or -token")
			os.Exit(1)
		}
		var err error
		jwtToken, err = loginOrRegister(*addr, *username, *password)
		if err != nil {
			fmt.Printf("auth failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("logged in as %s\n", *username)
	}

	conn, _, err := websocket.DefaultDialer.Dial("ws://"+*addr+"/ws", nil)
	if err != nil {
		fmt.Printf("connect failed: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	// 发送鉴权包
	sendJSON(conn, proto.AuthMessage{MsgType: proto.MsgTypeAuth, Token: jwtToken})
	fmt.Println("connected. commands: s <uid> <text> | g <group_id> <text> | q")

	// 后台接收
	go func() {
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("\ndisconnected:", err)
				os.Exit(0)
			}
			var msg map[string]any
			json.Unmarshal(data, &msg)
			fmt.Printf("\n[recv] %s\n> ", formatMsg(msg))
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	var seq int64
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		parts := strings.SplitN(strings.TrimSpace(scanner.Text()), " ", 3)
		if len(parts) == 0 || parts[0] == "" {
			continue
		}
		switch parts[0] {
		case "q":
			return
		case "s":
			if len(parts) < 3 {
				fmt.Println("usage: s <uid> <text>")
				continue
			}
			toUID, _ := strconv.ParseInt(parts[1], 10, 64)
			seq++
			sendJSON(conn, proto.Message{
				MsgType: proto.MsgTypeText, TargetType: proto.TargetTypeSingle,
				ToID: toUID, Content: parts[2], Timestamp: time.Now().UnixMilli(), Seq: seq,
			})
		case "g":
			if len(parts) < 3 {
				fmt.Println("usage: g <group_id> <text>")
				continue
			}
			groupID, _ := strconv.ParseInt(parts[1], 10, 64)
			seq++
			sendJSON(conn, proto.Message{
				MsgType: proto.MsgTypeText, TargetType: proto.TargetTypeGroup,
				ToID: groupID, Content: parts[2], Timestamp: time.Now().UnixMilli(), Seq: seq,
			})
		default:
			fmt.Println("unknown command")
		}
	}
}

func loginOrRegister(addr, username, password string) (string, error) {
	body, _ := json.Marshal(map[string]string{"username": username, "password": password})

	// 先尝试登录
	token, err := postJSON("http://"+addr+"/api/user/login", body)
	if err == nil {
		return token, nil
	}
	// 登录失败则注册再登录
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
		return "", fmt.Errorf("status %d: %s", resp.StatusCode, data)
	}
	var result map[string]string
	json.Unmarshal(data, &result)
	return result["token"], nil
}

func sendJSON(conn *websocket.Conn, v any) {
	data, _ := json.Marshal(v)
	conn.WriteMessage(websocket.TextMessage, data)
}

func formatMsg(msg map[string]any) string {
	msgType := int(toFloat(msg["msg_type"]))
	switch msgType {
	case proto.MsgTypeAck:
		return fmt.Sprintf("ACK seq=%v msg_id=%v", msg["seq"], msg["msg_id"])
	case proto.MsgTypeText:
		return fmt.Sprintf("from=%v content=%v", msg["from_uid"], msg["content"])
	default:
		return fmt.Sprintf("%v", msg)
	}
}

func toFloat(v any) float64 {
	f, _ := strconv.ParseFloat(fmt.Sprintf("%v", v), 64)
	return f
}
