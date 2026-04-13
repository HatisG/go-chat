package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// 配置参数，通过命令行传入
var (
	concurrency = flag.Int("c", 100, "并发连接数")
	messageRate = flag.Int("r", 1, "每个连接每秒发送消息数")
	duration    = flag.Int("d", 10, "压测持续时间（秒）")
	baseURL     = flag.String("url", "http://localhost:8080", "服务地址")
	toUserID    = flag.Int("to", 3, "消息接收方用户ID")
)

// 统计指标
var (
	connSuccess int64
	connFail    int64
	msgSent     int64
	msgReceived int64
	msgFail     int64
)

func main() {
	flag.Parse()

	log.Printf("=== 开始压测 ===")
	log.Printf("目标地址: %s", *baseURL)
	log.Printf("并发连接: %d", *concurrency)
	log.Printf("消息频率: %d msg/s/conn", *messageRate)
	log.Printf("持续时间: %d 秒", *duration)

	var wg sync.WaitGroup

	// 启动压测协程
	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			runClient(index)
		}(i)
		time.Sleep(5 * time.Millisecond) // 控制连接建立速率
	}

	// 定时打印统计信息
	ticker := time.NewTicker(2 * time.Second)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-ticker.C:
				printStats()
			case <-done:
				return
			}
		}
	}()

	wg.Wait()
	done <- true

	log.Printf("=== 压测结束 ===")
	printStats()
}

func runClient(id int) {
	// 1. 注册/登录 (简化起见，使用固定用户或动态生成)
	username := fmt.Sprintf("testuser_%d", id)
	token, _ := login(username)

	// 2. 建立 WebSocket 连接
	url := fmt.Sprintf("ws://localhost:8080/api/v1/chat/ws?token=%s", token)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		atomic.AddInt64(&connFail, 1)
		log.Printf("[Client %d] 连接失败: %v", id, err)
		return
	}
	atomic.AddInt64(&connSuccess, 1)
	defer conn.Close()

	// 3. 启动接收消息协程
	go func() {
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				return
			}
			atomic.AddInt64(&msgReceived, 1)
		}
	}()

	//4. 按照频率发送消息
	ticker := time.NewTicker(time.Second / time.Duration(*messageRate))
	defer ticker.Stop()

	deadline := time.Now().Add(time.Duration(*duration) * time.Second)

	for time.Now().Before(deadline) {
		<-ticker.C

		msg := map[string]interface{}{
			"type":       "chat",
			"to_user_id": *toUserID,
			"content":    fmt.Sprintf("benchmark message from %d", id),
		}
		msgBytes, _ := json.Marshal(msg)

		if err := conn.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
			atomic.AddInt64(&msgFail, 1)
			return
		}
		atomic.AddInt64(&msgSent, 1)
	}
}

// login 模拟用户登录，返回 token 和 userID
func login(username string) (string, uint) {
	// 尝试注册（忽略错误）
	regBody, _ := json.Marshal(map[string]string{
		"username": username,
		"password": "123456",
	})
	http.Post(*baseURL+"/api/v1/user/register", "application/json", bytes.NewReader(regBody))

	// 登录
	loginBody, _ := json.Marshal(map[string]string{
		"username": username,
		"password": "123456",
	})
	resp, err := http.Post(*baseURL+"/api/v1/user/login", "application/json", bytes.NewReader(loginBody))
	if err != nil {
		log.Fatalf("登录失败: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			Token string `json:"token"`
			ID    uint   `json:"id"`
		} `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	return result.Data.Token, result.Data.ID
}

func printStats() {
	log.Printf("连接: 成功=%d 失败=%d | 消息: 发送=%d 接收=%d 失败=%d",
		atomic.LoadInt64(&connSuccess),
		atomic.LoadInt64(&connFail),
		atomic.LoadInt64(&msgSent),
		atomic.LoadInt64(&msgReceived),
		atomic.LoadInt64(&msgFail),
	)
}
