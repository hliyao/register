package grpclb

import (
	"github.com/coreos/etcd/clientv3"
	clientv3naming "github.com/coreos/etcd/clientv3/naming"
	"golang.org/x/net/context"
	"google.golang.org/grpc/naming"

	"encoding/json"
	"errors"
	"os"
	"path"
	"sync"
	"time"

	"log"
	"register/pkg/config"
	"strconv"
)

const (
	ttl = int64(15)
)

var (
	errGRPCResolverNotReady   = errors.New("gRPC resolver is not ready.")
	defaultEtcdv3Resolver     *etcdv3Resolver
	defaultEtcdv3ResolverOnce sync.Once
)

func DefaultEtcdv3Resolver() *etcdv3Resolver {
	defaultEtcdv3ResolverOnce.Do(func() {
		const etcdClientConf = "etcd_client.json"
		clientConfigPath := os.Getenv(config.ClientConfigPathEnvName)
		if len(clientConfigPath) == 0 {
			clientConfigPath = config.DefaultClientConfigPath()
		}
		conf := path.Join(clientConfigPath, etcdClientConf)
		defaultEtcdv3Resolver = NewEtcdv3Resolver(context.Background(), conf)
	})

	return defaultEtcdv3Resolver
}

func GetEtcdv3Endpoints() ([]string, error) {
	const etcdClientConf = "etcd_client.json"
	clientConfigPath := os.Getenv(config.ClientConfigPathEnvName)
	if len(clientConfigPath) == 0 {
		clientConfigPath = config.DefaultClientConfigPath()
	}
	var endPoints []string
	conf := path.Join(clientConfigPath, etcdClientConf)
	f, err := os.Open(conf)
	if err != nil {
		log.Fatalf("Failed to open gRPC resolver config: %s, err: %v", conf, err)
		return endPoints, err
	}

	var cfg clientv3.Config
	if err = json.NewDecoder(f).Decode(&cfg); err != nil {
		log.Fatalf("Failed to decode gRPC resolver config: %s, err: %v", conf, err)
		f.Close()
		return endPoints, err
	}
	f.Close()

	return cfg.Endpoints, nil
}

type etcdv3Resolver struct {
	gr          *clientv3naming.GRPCResolver
	mu          sync.Mutex
	ready       chan struct{}
	serverWatch *ServerWatcherT
	myAddr      string
}

// NewEtcdv3Resolver creates a new resolver implemented with etcd client v3
func NewEtcdv3Resolver(ctx context.Context, conf string) *etcdv3Resolver {
	r := &etcdv3Resolver{}
	go r.initGRPCResolver(ctx, conf)
	return r
}

func (r *etcdv3Resolver) initGRPCResolver(ctx context.Context, conf string) {
	for {
		for {
			var cfg clientv3.Config
			var cli *clientv3.Client
			f, err := os.Open(conf)
			if err != nil {
				log.Fatalf("Failed to open gRPC resolver config: %s, err: %v", conf, err)
				break
			}

			if err = json.NewDecoder(f).Decode(&cfg); err != nil {
				log.Fatalf("Failed to decode gRPC resolver config: %s, err: %v", conf, err)
				f.Close()
				break
			}
			f.Close()

			//log.Debugf("Initializing gRPC resolver from config: %+v %+v", conf, cfg)
			cli, err = clientv3.New(cfg)
			if err != nil {
				log.Fatalf("Failed to create gRPC resolver from config: %+v, err: %v", cfg, err)
				break
			}

			r.mu.Lock()
			r.gr = &clientv3naming.GRPCResolver{Client: cli}
			if r.ready != nil {
				close(r.ready)
			}
			r.mu.Unlock()
			//log.Debugf("Initialized gRPC resolver from config: %+v %+v", conf, cfg)
			return
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(3 * time.Second):
		}
	}
}

// Register a named server to etcd
func (r *etcdv3Resolver) Register(ctx context.Context, name string, addr string) {
	r.myAddr = addr
	r.serverWatch = &ServerWatcherT{
		name:   name,
		ctxend: make(chan struct{}),
	}
	go func() {
		for {
			gr, err := r.wait(ctx, false)
			if err != nil {
				if ctx.Err() != nil {
					// context was cancelled or deadline exceeded
					break
				}
			}

			strIndex, bDo := r.registerIndex(ctx, gr, name)
			if !bDo {
				continue
			}

			leaseGrant, err := gr.Client.Grant(ctx, ttl)
			if err != nil {
				continue
			}

			err = gr.Update(ctx, name, naming.Update{Op: naming.Add, Addr: addr, Metadata: strIndex}, clientv3.WithLease(leaseGrant.ID))
			if err != nil {
				continue
			}

			r.serverWatch.resolver = gr
			go r.serverWatch.StartWatching(ctx)
			// KeepAlive until ctx closed or timeout
			respChan, err := gr.Client.KeepAlive(ctx, leaseGrant.ID)
			if err != nil {
				continue
			}

			for range respChan {
			}

			if ctx.Err() != nil {
				// context was cancelled or deadline exceeded
				log.Printf("Register End, err:%s", ctx.Err().Error())
				r.serverWatch.closeCtx()
				break
			}
		}
	}()

	return
}

func (r *etcdv3Resolver) registerIndex(ctx context.Context, gr *clientv3naming.GRPCResolver, name string) (string, bool) {
	bFirst := false
	nameIndexKey := "services/" + name + "/index"
	var compareValue string
	var nIndex int
	getResp, err := r.gr.Client.Get(ctx, nameIndexKey)
	//fmt.Printf("get key:%s,%+v.%+v", nameIndexKey, getResp, err)
	if err != nil {
		return "", false
	} else {
		if len(getResp.Kvs) == 0 {
			nIndex = 0
			bFirst = true
		} else {
			compareValue = string(getResp.Kvs[0].Value)
			/*if compareValue == "" {
				compareValue = "0"
			}*/
			log.Println("registerIndex compareValue:%s", compareValue)
			nIndex, err = strconv.Atoi(compareValue)
			if err != nil {
				nIndex = 0
				log.Fatalf("strconv.Atoi, err: %s, compareValue:%s", err.Error(), compareValue)
			}
		}
	}
	nIndex++
	if nIndex >= 80000000 {
		nIndex = 0 //数值非常大时重置
	}
	strIndex := strconv.Itoa(nIndex)
	if bFirst {
		_, err := gr.Client.Put(ctx, nameIndexKey, strIndex)
		if err != nil {
			return "", false
		}
	} else {
		txn := gr.Client.Txn(ctx)
		txnResp, err := txn.If(clientv3.Compare(clientv3.Value(nameIndexKey), "=", compareValue)).
			Then(clientv3.OpPut(nameIndexKey, strIndex)).
			Commit()
		if txnResp.Succeeded {
			log.Printf("Etcd Txn SetIndex Suc, key:%s, value:%d", nameIndexKey, nIndex)
		} else {
			log.Fatalf("Etcd Txn SetIndex Fail, err:%+v, key:%s, value:%d", err, nameIndexKey, nIndex)
			return "", false
		}
	}
	return strIndex, true
}

// Unregister a named server from etcd
func (r *etcdv3Resolver) Unregister(ctx context.Context, name string, addr string) error {
	if gr, err := r.wait(ctx, true); err != nil {
		return err
	} else {
		return gr.Update(ctx, name, naming.Update{Op: naming.Delete, Addr: addr})
	}
}

func (r *etcdv3Resolver) Resolve(target string) (naming.Watcher, error) {
	if gr, err := r.wait(context.Background(), false); err != nil {
		return nil, err
	} else {
		return gr.Resolve(target)
	}
}

func (r *etcdv3Resolver) wait(ctx context.Context, failFast bool) (*clientv3naming.GRPCResolver, error) {
	for {
		r.mu.Lock()
		// already initialized
		if r.gr != nil {
			r.mu.Unlock()
			return r.gr, nil
		}

		// r.gr == nil
		if failFast {
			r.mu.Unlock()
			return nil, errGRPCResolverNotReady
		}

		ready := r.ready
		if ready == nil {
			ready = make(chan struct{})
			r.ready = ready
		}
		r.mu.Unlock()

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ready:
			// continue
		}
	}
}

type ProcessInfo struct {
	addr  string
	index uint32
}

type ProcessIndexT struct {
	index uint32
}

type ServerWatcherT struct {
	name         string
	resolver     naming.Resolver
	processes    []ProcessInfo
	processMutex sync.RWMutex
	ctxend       chan struct{}
}

func (sw *ServerWatcherT) closeCtx() {
	close(sw.ctxend)
}

func (sw *ServerWatcherT) StartWatching(ctx context.Context) {
	for {
		log.Printf("StartWatching")
		if ctx.Err() != nil {
			// context was cancelled or deadline exceeded
			break
		}

		w, err := sw.resolver.Resolve(sw.name)
		if err != nil {
			log.Fatalf("[watcher] Failed to resolve %s, err: %v, try later", sw.name, err)
			time.Sleep(time.Second * 3)
			continue
		}

		for {
			updates, err := w.Next()
			if err != nil {
				log.Fatalf("[watcher] Next err: %v", err)
				// re-resolve in out loop
				break
			}
			sw.onUpdates(updates)
			select {
			case <-sw.ctxend:
				log.Printf("StartWatching End, ctx err:%+v", ctx.Err())
				return
			case <-time.After(1 * time.Millisecond):
			}
		}
	}
}

//定时调度，不用主动通知
func (sw *ServerWatcherT) onUpdates(updates []*naming.Update) {
	sw.processMutex.Lock()
	for _, update := range updates {
		switch update.Op {
		case naming.Add:
			// remove old addr
			for i, v := range sw.processes {
				if v.addr == update.Addr {
					copy(sw.processes[i:], sw.processes[i+1:])
					sw.processes = sw.processes[:len(sw.processes)-1]
					log.Println("[watcher] %s delete existing address addr %s", sw.name, update.Addr)
					break
				}
			}

			//log.Infof("[watcher] %s add new addr %+v", sw.name, update)
			proncessIndex := 0
			//ok := false
			if update.Metadata == nil {
				continue
			} else {
				strIndex, ok := update.Metadata.(string)
				if !ok {
					continue
				}
				var err error
				proncessIndex, err = strconv.Atoi(strIndex)
				if err != nil {
					continue
				}
			}
			sw.processes = append(sw.processes, ProcessInfo{addr: update.Addr, index: uint32(proncessIndex)})
			log.Printf("[watcher] %s add new addr %s, index:%d", sw.name, update.Addr, proncessIndex)
		case naming.Delete:
			for i, v := range sw.processes {
				if v.addr == update.Addr {
					copy(sw.processes[i:], sw.processes[i+1:])
					sw.processes = sw.processes[:len(sw.processes)-1]
					log.Printf("[watcher] %s delete addr %s", sw.name, update.Addr)
					break
				}
			}
		default:
			log.Fatalln("[watcher] Unknown update.Op ", update.Op)
		}
	}
	sw.processMutex.Unlock()
}

func (sw *ServerWatcherT) GetProcesses() []ProcessInfo {
	defer sw.processMutex.RUnlock()
	sw.processMutex.RLock()
	nLen := len(sw.processes)
	if nLen == 0 {
		return nil
	}
	newProcessInfo := make([]ProcessInfo, nLen)
	for i := 0; i < nLen; i++ {
		newProcessInfo[i] = sw.processes[i]
	}
	return newProcessInfo
}

