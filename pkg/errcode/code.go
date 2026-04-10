package errcode

const (
	Success       = 0
	InvalidParams = 400
	Unauthorized  = 401
	ServerError   = 500

	// 用户相关 (10000-19999)
	UserNotFound      = 10001
	UserAlreadyExists = 10002
	UserPasswordWrong = 10003
)

var MsgMap = map[int]string{
	Success:           "success",
	InvalidParams:     "请求参数错误",
	Unauthorized:      "未授权",
	ServerError:       "服务器内部错误",
	UserNotFound:      "用户不存在",
	UserAlreadyExists: "用户名已存在",
	UserPasswordWrong: "密码错误",
}

func GetMsg(code int) string {
	if msg, ok := MsgMap[code]; ok {
		return msg
	}
	return "未知错误"
}
