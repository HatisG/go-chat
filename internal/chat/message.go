package chat

// ChatMessage 消息结构体（用于 MQ 传输）
type ChatMessage struct {
	FromUserID uint   `json:"from_user_id"`
	ToUserID   uint   `json:"to_user_id"`
	Content    string `json:"content"`
	MsgType    string `json:"msg_type"`
}
