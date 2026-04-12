package friend

type FriendInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Nickname string `jspn:"nickname"`
	Avatar   string `json:"avatar"`
	Status   int    `json:"status"`
}
