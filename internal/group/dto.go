package group

// CreateGroupReq 创建群请求
type CreateGroupReq struct {
	Name string `json:"name" binding:"required,max=64"`
}

// GroupInfo 群信息（返回给前端）
type GroupInfo struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Avatar      string `json:"avatar"`
	CreatorID   uint   `json:"creator_id"`
	MemberCount int64  `json:"member_count"`
	Role        int    `json:"role"` // 当前用户在群中的角色
}

// MemberInfo 成员信息
type MemberInfo struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Role     int    `json:"role"`
}

type GroupMessageResp struct {
	ID           uint   `json:"id"`
	GroupID      uint   `json:"group_id"`
	FromUserID   uint   `json:"from_user_id"`
	FromUserName string `json:"from_user_name"`
	Content      string `json:"content"`
	MsgType      string `json:"msg_type"`
	CreatedAt    uint   `json:"created_at"`
}
