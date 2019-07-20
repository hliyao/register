package cpu

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

var preCputime, preGetCputime int64
var cpuUsage float64

const hertz = 100

// GetCpu 返回CPU使用率
func GetCpuUsage() float64 {
	return cpuUsage
}

// 计算 CPU 使用率
func calCpuUsage() {
	var curCputime, curGetCputime int64
	pid := os.Getpid()
	for {
		curCputime, curGetCputime = cputime(pid)
		if preGetCputime != 0 {
			cpuUsage = float64((curCputime-preCputime)*1000000000) / float64(hertz*(curGetCputime-preGetCputime))
		}
		preCputime, preGetCputime = curCputime, curGetCputime
		time.Sleep(time.Second * 5)
	}
}

// 计算 CPU的使用总数
func cputime(pid int) (cputime, curtime int64) {
	// 参考文档
	// http://stackoverflow.com/questions/16726779/how-do-i-get-the-total-cpu-usage-of-an-application-from-proc-pid-stat
	bin, err := ioutil.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return
	}
	curtime = time.Now().UnixNano()
	arr := strings.Split(strings.TrimSpace(string(bin)), " ")
	utime, _ := strconv.ParseInt(arr[13], 10, 64)
	stime, _ := strconv.ParseInt(arr[14], 10, 64)
	cutime, _ := strconv.ParseInt(arr[15], 10, 64)
	cstime, _ := strconv.ParseInt(arr[16], 10, 64)
	totaltime := utime + stime + cutime + cstime
	return totaltime, curtime
}

func init() {
	go calCpuUsage()
}
