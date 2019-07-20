package monitor

import (
	"os"
	"strings"
)

var rule string

func UseRule(r string) {
	rule = r
}

//遍历模块，获取信息
func Status() (as AppStatus) {
	cmd := strings.Join(os.Args, " ")
	host, _ := os.Hostname()
	as.Name = host + ":" + cmd
	registerLock.RLock()
	for _, mod := range registerModules {
		modStat := mod.MonitorStatus()
		if len(modStat.Events) == 0 {
			modStat.Events = DefaultEvents()
		}
		as.Modules = append(as.Modules, mod.MonitorStatus())
	}
	registerLock.RUnlock()
	return as
}

func DefaultEvents() []string {
	return defaultEvents
}

func OKStatus(name string) ModuleStatus {
	return ModuleStatus{
		Name:   name,
		Status: STATUS_OK,
		Msg:    name + " is fine",
		Events: DefaultEvents(),
		Rule:   rule,
	}
}
