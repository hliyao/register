package monitor

//模块的状态
const (
	STATUS_OK       = "OK"
	STATUS_WARNING  = "WARNING"  //警告
	STATUS_CRITICAL = "CRITICAL" //危险
	STATUS_UNKNOWN  = "UNKNOWN"  //未知
	STATUS_TIMEOUT  = "TIMEOUT"  //超时
)

//通知方式
const (
	EVENT_SMS    = "sms"
	EVENT_MAIL   = "mail"
	EVENT_WEIXIN = "weixin"
)

var defaultEvents = []string{
	EVENT_MAIL,
	EVENT_SMS,
	EVENT_WEIXIN,
}

//实现该接口表示模块可监控的
type Monitored interface {
	//用于获取状态,当被调用时,返回模块的状态信息
	MonitorStatus() ModuleStatus
}

//模块的状态信息
type ModuleStatus struct {
	//模块名称
	Name string `json:"name"`
	//模块状态
	Status string `json:"status"`
	//状态信息
	Msg string `json:"msg"`
	//模块事件
	Events []string `json:"events"`
	Rule   string   `json:"rule,omitempty"`
}

//应用的状态信息
type AppStatus struct {
	//应用名称
	Name string `json:"name"`
	//各个模块信息
	Modules []ModuleStatus `json:"modules"`
}
