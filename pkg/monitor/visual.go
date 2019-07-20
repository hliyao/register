package monitor

import (
	"strings"
	"sync"
)

var mutex sync.Mutex
var visualMonitors []VisualMonitor

func VisualText() string {
	var texts []string
	for _, m := range visualMonitors {
		texts = append(texts, "\""+m.Text()+"\"")
	}
	return "[" + strings.Join(texts, ",") + "]"
}

func ResetVisualMonitor() {
	visualMonitors = make([]VisualMonitor, 0, len(visualMonitors))
}

func SetVisualMonitorString(s string) {
	ResetVisualMonitor()
	RegistVisualMonitorString(s)
}

func RegistVisualMonitorString(s string) {
	defer mutex.Unlock()
	mutex.Lock()
	visualMonitors = append(visualMonitors, &FuncVisualMonitor{func() string { return s }})
}

func SetVisualMonitorFunc(f func() string) {
	ResetVisualMonitor()
	RegistVisualMonitorFunc(f)
}

func RegistVisualMonitorFunc(f func() string) {
	defer mutex.Unlock()
	mutex.Lock()
	visualMonitors = append(visualMonitors, &FuncVisualMonitor{f})
}

func SetVisualMonitor(visual VisualMonitor) {
	ResetVisualMonitor()
	RegistVisualMonitor(visual)
}

func RegistVisualMonitor(visual VisualMonitor) {
	defer mutex.Unlock()
	mutex.Lock()
	visualMonitors = append(visualMonitors, visual)
}

type VisualMonitor interface {
	Text() string
}

type FuncVisualMonitor struct {
	Func func() string
}

func (this *FuncVisualMonitor) Text() string {
	return this.Func()
}
