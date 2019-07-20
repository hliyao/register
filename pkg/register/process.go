package grpclb

import (
	"gitlab.ttyuyin.com/golang/gudetama/log"
	"sync"
)

var (
	processOnce     sync.Once
	etcdProcessInst *EtcdProcessT
)

func GetEtcdProcess() *EtcdProcessT {
	processOnce.Do(func() {
		etcdProcessInst = &EtcdProcessT{}
	})
	return etcdProcessInst
}

type EtcdProcessT struct {
	defaultEtcdv3Resolver *etcdv3Resolver
}

func (e *EtcdProcessT) SetResolve(resolve *etcdv3Resolver) {
	defaultEtcdv3Resolver = resolve
}

func (e *EtcdProcessT) IsSmallProcess() bool {
	if defaultEtcdv3Resolver == nil {
		log.Warnln("IsSmallProcess defaultEtcdv3Resolver nil")
		return false
	} else if defaultEtcdv3Resolver.myAddr == "" {
		log.Warnln("IsSmallProcess defaultEtcdv3Resolver.myAddr nil")
		return false
	}
	processInfos := defaultEtcdv3Resolver.serverWatch.GetProcesses()
	if processInfos == nil {
		log.Errorf("IsSmallProcess GetProcesses cnt 0, name:%s", defaultEtcdv3Resolver.serverWatch.name)
		return false
	}
	var smallI int
	var smallIndex uint32
	for i, info := range processInfos {
		if i == 0 {
			smallIndex = info.index
			smallI = i
		} else {
			if info.index < smallIndex {
				smallIndex = info.index
				smallI = i
			}
		}
	}
	processInfo := processInfos[smallI]
	if processInfo.addr == defaultEtcdv3Resolver.myAddr {
		log.Infof("IsSmallProcess addr:%s, name:%s", defaultEtcdv3Resolver.myAddr, defaultEtcdv3Resolver.serverWatch.name)
		return true
	}
	return false
}

func (e *EtcdProcessT) IsBigProcess() bool {
	if defaultEtcdv3Resolver == nil {
		log.Warnln("IsSmallProcess defaultEtcdv3Resolver nil")
		return false
	} else if defaultEtcdv3Resolver.myAddr == "" {
		log.Warnln("IsSmallProcess defaultEtcdv3Resolver.myAddr nil")
		return false
	}
	processInfos := defaultEtcdv3Resolver.serverWatch.GetProcesses()
	if processInfos == nil {
		log.Errorf("IsBigProcess GetProcesses cnt 0, name:%s", defaultEtcdv3Resolver.serverWatch.name)
		return false
	}
	var bigI int
	var bigIndex uint32
	for i, info := range processInfos {
		if i == 0 {
			bigIndex = info.index
			bigI = i
		} else {
			if info.index > bigIndex {
				bigIndex = info.index
				bigI = i
			}
		}
	}
	processInfo := processInfos[bigI]
	if processInfo.addr == defaultEtcdv3Resolver.myAddr {
		log.Infof("IsSmallProcess addr:%s, name:%s", defaultEtcdv3Resolver.myAddr, defaultEtcdv3Resolver.serverWatch.name)
		return true
	}
	return false
}
