package grpc

var (
	// Common Error
	Success        = 0 // 成功
	ErrSys               = -2 // 系统错误
	ErrDatabaseOperation = -3
	ErrSystemBusy        = -4 // 系统繁忙
	ErrRequestTimeout    = -5 // 处理超时
	ErrBadRequest        = -6 // 无效的请求
	ErrTrafficAdmin      = -8 // 流量管制
)

// MessageFromCode get message associated with the code
func MessageFromCode(code int) string {
	if m, ok := CodeMessageMap[code]; ok {
		return m
	}

	return ""
}

var CodeMessageMap = map[int]string{
	Success:           "成功",
	ErrSys:            "系统错误",
	ErrSystemBusy:     "系统繁忙",
	ErrRequestTimeout: "处理超时",
	ErrBadRequest:     "无效的请求",
}
