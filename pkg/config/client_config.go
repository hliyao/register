package config

import (
	"fmt"
	"math/rand"
	"time"

	"os"
	"sync"
)

// Endpoint ...
type Endpoint struct {
	IP        string `json:"ip"`
	Port      int    `json:"port"`
	SectBegin int    `json:"sect_begin"`
	SectEnd   int    `json:"sect_end"`
}

// Description ...
func (ep *Endpoint) String() string {
	return fmt.Sprintf("%s:%d", ep.IP, ep.Port)
}

//===========================================================================================
const (
	defaultConnectTimeout = 300
	defaultSocketTimeout  = 5000
)

// ClientConfig ...
type ClientConfig struct {
	endpoints      []*Endpoint `json:"endpoints"`
	packageName    string      `json:"package_name"`
	connectTimeout int         `json:"connect_timeout"` // in microseconds
	socketTimeout  int         `json:"socket_timeout"`  // in microseconds

	mu         sync.RWMutex
	adapter    string
	filePath   string
	modTime    time.Time
	updateTime time.Time
}

// NewClientConfig ...
func NewClientConfig(filePath string, adapter ...string) *ClientConfig {
	cc := new(ClientConfig)
	cc.connectTimeout = defaultConnectTimeout
	cc.socketTimeout = defaultSocketTimeout

	var adapterName = "ini"
	if len(adapter) > 0 {
		adapterName = adapter[0]
	}

	cc.filePath = filePath
	cc.adapter = adapterName

	return cc
}

// ConnectTimeout getter of connectTimeout
func (cc *ClientConfig) ConnectTimeout() time.Duration {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return time.Duration(cc.connectTimeout) * time.Millisecond
}

// SocketTimeout getter of socketTimeout
func (cc *ClientConfig) SocketTimeout() time.Duration {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return time.Duration(cc.socketTimeout) * time.Millisecond
}

// PackageName getter of packageName
func (cc *ClientConfig) PackageName() string {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return cc.packageName
}

// Endpoints getter of endpoints
func (cc *ClientConfig) Endpoints() []*Endpoint {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return cc.endpoints
}

// GetRandomEndpoint get a random point
func (cc *ClientConfig) GetRandomEndpoint() *Endpoint {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	var endpoint *Endpoint
	if len(cc.endpoints) > 0 {
		randomGenerator := rand.New(rand.NewSource(time.Now().UnixNano()))
		endpoint = cc.endpoints[randomGenerator.Intn(len(cc.endpoints))]
	}

	return endpoint
}

func (cc *ClientConfig) reloadNoLock() error {
	finfo, err := os.Stat(cc.filePath)
	if err != nil {
		return err
	}

	// file was not modified
	if finfo.ModTime() == cc.modTime {
		return nil
	}

	conf, err := NewConfig(cc.adapter, cc.filePath)
	if err != nil {
		return err
	}

	connectTimeout := conf.DefaultInt("ClientTimeout::ConnectTimeoutMS", defaultConnectTimeout)
	socketTimeout := conf.DefaultInt("ClientTimeout::SocketTimeoutMS", defaultSocketTimeout)
	serverCount := conf.DefaultInt("Server::ServerCount", 0)
	packageName := conf.DefaultString("Server::PackageName", "Unknown")

	endpoints := make([]*Endpoint, 0, serverCount)
	for i := 0; i < serverCount; i++ {
		sectionName := fmt.Sprintf("Server%d", i)
		ip := conf.String(fmt.Sprintf("%s::SVR_IP", sectionName))
		port := conf.DefaultInt(fmt.Sprintf("%s::SVR_Port", sectionName), 0)
		sectBegin := conf.DefaultInt(fmt.Sprintf("%s::Sect_Begin", sectionName), 0)
		sectEnd := conf.DefaultInt(fmt.Sprintf("%s::Sect_End", sectionName), 0)
		if ip != "" && port != 0 {
			ep := &Endpoint{
				IP:        ip,
				Port:      port,
				SectBegin: sectBegin,
				SectEnd:   sectEnd,
			}
			endpoints = append(endpoints, ep)
		}
	}

	cc.socketTimeout = socketTimeout
	cc.connectTimeout = connectTimeout
	cc.packageName = packageName
	cc.endpoints = endpoints
	cc.updateTime = time.Now()
	cc.modTime = finfo.ModTime()

	return nil
}

func (cc *ClientConfig) Reload(interval time.Duration) {
	cc.mu.RLock()
	if time.Since(cc.updateTime) < interval {
		// no need to reload
		cc.mu.RUnlock()
		return
	}
	cc.mu.RUnlock()

	cc.mu.Lock()
	defer cc.mu.Unlock()

	if time.Since(cc.updateTime) < interval {
		// check again
		return
	}

	cc.reloadNoLock()
}
