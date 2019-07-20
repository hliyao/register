package web

import (
	"bytes"
	"fmt"
	"golang.52tt.com/pkg/versioning/svn"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"

	"golang.52tt.com/pkg/monitor"
	"golang.52tt.com/pkg/monitor/cpu"
)

var typeName []string = []string{"Bytes", "KB", "MB", "GB", "TB"}
var upTimeStr string
var currentQps float64
var perTotalCount int64
var versionInfo []string
var monitorCount int64

const (
	qpsInterval = 10
)

// 将大小转化为更可读的形式
func toHumanSize(siteBtyes int64) string {
	siteBtyesF := float64(siteBtyes)
	i := 0
	for ; siteBtyesF > 1024; i++ {
		siteBtyesF = siteBtyesF / 1024
	}
	//在线上Linux服务器差一个量级的单位
	if i > 4 {
		i = 4
	}
	return fmt.Sprintf("%.1f%s", siteBtyesF, typeName[i])
}

// 获取进程当前的内存占用大小
func getMem() string {
	var mem int64
	statm := "/proc/" + strconv.Itoa(os.Getpid()) + "/statm"
	dat, err := ioutil.ReadFile(statm)
	if err == nil {
		stats := bytes.Split(dat, []byte(" "))
		mem, _ = strconv.ParseInt(string(stats[1]), 10, 64)
		mem *= 4 * 1024
	}
	return toHumanSize(mem)
}

type BaseVisualMonitor string

func (v BaseVisualMonitor) Text() string {
	return string(v)
}

//监控handlers，在初始化时启动，提供url接口查询qps信息
func init() {
	upTimeStr = time.Now().Format("2006-01-02 15:04:05")
	//定期刷新监控的最新信息
	go func() {
		for {
			time.Sleep(qpsInterval * time.Second)
			currentQps = float64((monitorCount - perTotalCount)) /// float64(qpsInterval)
			perTotalCount = monitorCount
		}
	}()
}

func Incr(n int64) {
	atomic.AddInt64(&monitorCount, n)
	//fmt.Println("----------test cnt:", monitorCount)
}

func GetCurrentQPS() float64 {
	return currentQps
}

func MonitorHandle(w http.ResponseWriter, r *http.Request) {
	mutexCnt, _ := runtime.MutexProfile(nil)
	blockCnt, _ := runtime.BlockProfile(nil)
	goroutineCnt := runtime.NumGoroutine()
	var memStatus runtime.MemStats
	runtime.ReadMemStats(&memStatus)
	sys := strconv.FormatInt(int64(memStatus.Sys/1024/1024), 10)
	heapSys := strconv.FormatInt(int64(memStatus.HeapSys/1024/1024), 10)
	totalAlloc := strconv.FormatInt(int64(memStatus.TotalAlloc/1024/1024), 10)
	heapInuse := strconv.FormatInt(int64(memStatus.HeapInuse/1024/1024), 10)
	heapAlloc := strconv.FormatInt(int64(memStatus.HeapAlloc/1024/1024), 10)

	w.Write([]byte(fmt.Sprintf(`{"cpu":%.f,"mem":"%s","ver":"%s","up_time":"%s","total_count":%d,"last_10_second_qps":%f,"gorotine_info":"%s","heap_info":"%s","text":%s}`,
		cpu.GetCpuUsage()*100,
		getMem(),
		runtime.Version()+" "+svn.CodeRevision,
		upTimeStr,
		monitorCount,
		currentQps,
		"groutine:"+strconv.Itoa(goroutineCnt)+",block:"+strconv.Itoa(blockCnt)+",mutex:"+strconv.Itoa(mutexCnt),
		"sys:"+sys+",heapsys:"+heapSys+",allalloc:"+totalAlloc+",heapinuse:"+heapInuse+",heapalloc:"+heapAlloc,
		monitor.VisualText(),
	)))
}
