package monitor

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

var splitTag = "-----"

func RecordErr(moduleName string, err string) {
	defaultErrMonitor.Record(moduleName, err)
}

func RecordMailErr(moduleName string, err string) {
	defaultErrMailMonitor.Record(moduleName, err)
}

// 记录UpdateCost失败， 进行报警
type ErrMonitor struct {
	name     string
	events   []string
	errors   map[string]int64
	lock     sync.Mutex
	lastTime int64
}

func (this *ErrMonitor) MonitorStatus() ModuleStatus {
	if this.errors == nil || len(this.errors) == 0 {
		return OKStatus(this.name)
	}
	this.lock.Lock()
	defer this.lock.Unlock()

	msg := ""
	errors := this.errors
	for key, times := range errors {
		arr := strings.Split(key, splitTag)
		msg = msg + fmt.Sprintf("\n [block: %s] \n [error:\n%s ] \n [times : %d] \n =========================\n", arr[0], arr[1], times)
	}

	if now := time.Now().Unix(); now-this.lastTime > 60 {
		this.errors = make(map[string]int64)
		this.lastTime = now
	}

	events := this.events
	if len(events) == 0 {
		events = DefaultEvents()
	}
	return ModuleStatus{
		Name:   this.name,
		Status: STATUS_CRITICAL,
		Msg:    msg,
		Rule:   rule,
		Events: events,
	}
}
func (this *ErrMonitor) Record(moduleName string, err string) {
	this.lock.Lock()
	defer this.lock.Unlock()
	now := time.Now().Unix()
	if this.errors == nil || now-this.lastTime > 30 {
		this.errors = map[string]int64{}
		this.lastTime = now
	}
	key := moduleName + splitTag + err
	if times, ok := this.errors[key]; ok {
		this.errors[key] = times + 1
	} else {
		this.errors[key] = 1
	}
}
