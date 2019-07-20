package monitor

import (
	"fmt"
	"sync"
	"time"
)

var panicMonitor *PanicMonitor

/*func init() {
	panicMonitor = &PanicMonitor{
		lastTime: time.Now().Unix(),
	}
	RegisterMonitor(panicMonitor)
}*/

func RecordPanic(err string) {
	panicMonitor.Record(err)
}

// 记录 PANIC 进行报警
type PanicMonitor struct {
	errors   map[string]int64
	lock     sync.Mutex
	lastTime int64
}

func (this *PanicMonitor) MonitorStatus() ModuleStatus {
	if this.errors == nil || len(this.errors) == 0 {
		return OKStatus("panic_msg")
	}
	this.lock.Lock()
	defer this.lock.Unlock()

	msg := ""
	errors := this.errors
	for name, times := range errors {
		msg = msg + fmt.Sprintf("[error:\n%s ] \n [times : %d] \n =========================\n", name, times)
	}

	if now := time.Now().Unix(); now-this.lastTime > 60 {
		this.errors = make(map[string]int64)
		this.lastTime = now
	}

	return ModuleStatus{
		Name:   "panic_msg",
		Status: STATUS_CRITICAL,
		Msg:    msg,
		Rule:   rule,
		Events: []string{EVENT_MAIL},
	}
}
func (this *PanicMonitor) Record(err string) {
	this.lock.Lock()
	defer this.lock.Unlock()
	now := time.Now().Unix()
	if this.errors == nil || now-this.lastTime > 30 {
		this.errors = map[string]int64{}
		this.lastTime = now
	}
	if times, ok := this.errors[err]; ok {
		this.errors[err] = times + 1
	} else {
		this.errors[err] = 1
	}
}
