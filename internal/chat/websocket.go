package chat

import (
	"encoding/json"
	"go-chat/internal/cache"
	"go-chat/internal/group"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	//允许无消息最大时间
	pongWait = 60 * time.Second
	//ping时间间隔
	pingPeriod = (pongWait * 9) / 10
	//写入超时时间
	writeWait = 10 * time.Second
)

// http升级为websocket
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// client表示一个websocket连接
type Client struct {
	UserID  uint
	Conn    *websocket.Conn
	Send    chan []byte
	Hub     *Hub
	Service *Service
}

// Hub管理所有在线连接
type Hub struct {
	Clients    map[uint]*Client
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan *WSMessage
	mu         sync.RWMutex
}

// Message消息结构体
type WSMessage struct {
	Type     string `json:"type"` //chat,heartbeat,ack
	ToUserID uint   `json:"to_user_id"`
	GroupID  uint   `json:"group_id,omitempty"`
	MsgType  string `json:"msg_type"`
	Content  string `json:"content"`
	FileName string `json:"file_name,omitempty"`
}

// 创建hub
func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[uint]*Client),
		Register:   make(chan *Client, 100),
		Unregister: make(chan *Client, 100),
		Broadcast:  make(chan *WSMessage, 1000),
	}
}

// 启动hub主循环
func (h *Hub) Run() {
	log.Println("Hub.Run() 启动")
	for {
		select {
		case client := <-h.Register:
			log.Printf("收到 Register: 用户 %d", client.UserID)
			h.mu.Lock()
			h.Clients[client.UserID] = client
			h.mu.Unlock()
			log.Printf("用户 %d 上线,当前在线：%d", client.UserID, len(h.Clients))

		case client := <-h.Unregister:
			log.Printf("收到 Unregister: 用户 %d", client.UserID)
			h.mu.Lock()
			if _, ok := h.Clients[client.UserID]; ok {
				delete(h.Clients, client.UserID)
				close(client.Send)
			}
			h.mu.Unlock()
			log.Printf("用户 %d 下线,当前在线：%d", client.UserID, len(h.Clients))

		case msg := <-h.Broadcast:
			log.Printf("收到 Broadcast: To=%d", msg.ToUserID)
			h.mu.RLock()
			if client, ok := h.Clients[msg.ToUserID]; ok {
				msgBytes, _ := json.Marshal(msg)
				client.Send <- msgBytes
			}
			h.mu.RUnlock()
		}

	}

}

func ServerWS(hub *Hub, service *Service, userID uint, w http.ResponseWriter, r *http.Request) {

	//升级websocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("websocket升级失败:", err)
		return
	}

	//创建client
	client := &Client{
		UserID:  userID,
		Conn:    conn,
		Send:    make(chan []byte, 256),
		Hub:     hub,
		Service: service,
	}

	//注册到hub
	log.Printf("ServeWS: 准备注册用户 %d", userID)
	hub.Register <- client
	log.Printf("ServeWS: 用户 %d 注册成功", userID)

	//读写协程
	go pushOfflineMessage(userID, client)
	go client.ReadPump()
	go client.WritePump()
}

func (c *Client) ReadPump() {
	defer func() {
		log.Printf("ReadPump defer 执行: 用户 %d", c.UserID)
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	//初始时间
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))

	//pong处理器
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, data, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("用户 %d 连接异常关闭：%v", c.UserID, err)
			}
			break
		}

		var wsMsg WSMessage
		if err := json.Unmarshal(data, &wsMsg); err != nil {
			continue
		}

		switch wsMsg.Type {
		case "chat":
			err := c.Service.SendMessage(c.UserID, wsMsg.ToUserID, wsMsg.MsgType, wsMsg.Content)
			if err != nil {
				c.Send <- []byte(`{"type":"error","content":"` + err.Error() + `"}`)
				continue
			}
		case "group_chat":
			err := c.Service.SendGroupMessage(wsMsg.GroupID, c.UserID, wsMsg.MsgType, wsMsg.Content)
			if err != nil {
				c.Send <- []byte(`{"type":"error","content":"` + err.Error() + `"}`)
				continue
			}
		case "heartbeat":
			c.Send <- []byte(`{"type":"heartbeat"}`)
		}

	}

}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		}

	}

}

// 推送离线消息
func pushOfflineMessage(userID uint, client *Client) {
	messages, err := cache.GetOfflineMessage(userID)
	if err != nil || len(messages) == 0 {
		return
	}

	log.Printf("用户 %d 上线, 推送 %d 条离线消息", userID, len(messages))

	for _, msg := range messages {
		wsMsg := WSMessage{
			Type:     "chat",
			ToUserID: userID,
			Content:  msg.Content,
		}
		msgBytes, _ := json.Marshal(wsMsg)
		client.Send <- msgBytes

		time.Sleep(5 * time.Millisecond)
	}

}

func (h *Hub) SendToUser(msg interface{}) {
	// 类型断言，处理群消息
	if groupMsg, ok := msg.(*group.GroupWSMessage); ok {
		h.mu.RLock()
		client, online := h.Clients[groupMsg.ToUserID]
		h.mu.RUnlock()

		if online {
			msgBytes, _ := json.Marshal(groupMsg)
			select {
			case client.Send <- msgBytes:
			default:
			}
		}
	}
}
