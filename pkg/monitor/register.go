package monitor

import (
	"sync"
)

//被监控的模块
type MonitorModule struct {
	Monitored
	Name  string
	count int
}

var registerLock sync.RWMutex
var registerModules = make(map[string]*MonitorModule)

var defaultErrMonitor *ErrMonitor
var defaultErrMailMonitor *ErrMonitor

/*func init() {
	defaultErrMonitor = &ErrMonitor{name: "errmsg"}
	defaultErrMailMonitor = &ErrMonitor{name: "errmsg_mailonly", events: []string{EVENT_MAIL}}
	RegisterMonitor(defaultErrMonitor)
	RegisterMonitor(defaultErrMailMonitor)
}*/

//向监控程序注册模块,使得应用的该模块可以被监控
//@m:被监控的模块
func RegisterMonitor(m Monitored) error {
	registerLock.Lock()
	defer registerLock.Unlock()
	var name = m.MonitorStatus().Name
	if mod, ok := registerModules[name]; ok {
		mod.count++
	} else {
		registerModules[name] = &MonitorModule{
			Name:      name,
			count:     1,
			Monitored: m,
		}
	}
	return nil
}

//删除需要监控的模块
//@m:被监控的模块
func UnregisterMonitor(name string) error {
	registerLock.Lock()
	defer registerLock.Unlock()

	if mod, ok := registerModules[name]; ok {
		mod.count--
		if mod.count == 0 {
			delete(registerModules, name)
		}
	}
	return nil
}