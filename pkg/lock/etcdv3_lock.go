package lock

import (
	"encoding/json"
	"errors"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"golang.52tt.com/pkg/config"
	"golang.org/x/net/context"
	"os"
	"path"
	"time"

	"gitlab.ttyuyin.com/golang/gudetama/log"
)

const (
	ttl = 30
)

var (
	defaultEtcdv3Client *etcdv3Client
)

func InitEtcdv3Client() {
	const etcdClientConf = "etcd_client.json"
	clientConfigPath := os.Getenv(config.ClientConfigPathEnvName)
	if len(clientConfigPath) == 0 {
		clientConfigPath = config.DefaultClientConfigPath()
	}
	conf := path.Join(clientConfigPath, etcdClientConf)
	defaultEtcdv3Client = NewEtcdv3Client(context.Background(), conf)
}

type etcdv3Client struct {
	cli *clientv3.Client
}

// NewEtcdv3Resolver creates a new resolver implemented with etcd client v3
func NewEtcdv3Client(ctx context.Context, conf string) *etcdv3Client {
	r := &etcdv3Client{}
	go r.initEtcdClient(ctx, conf)
	return r
}

func (r *etcdv3Client) initEtcdClient(ctx context.Context, conf string) {
	for {
		for {
			var cfg clientv3.Config
			var cli *clientv3.Client
			f, err := os.Open(conf)
			if err != nil {
				log.Errorf("Failed to open gRPC resolver config: %s, err: %v", conf, err)
				break
			}

			if err = json.NewDecoder(f).Decode(&cfg); err != nil {
				log.Errorf("Failed to decode gRPC resolver config: %s, err: %v", conf, err)
				f.Close()
				break
			}
			f.Close()

			//log.Debugf("Initializing gRPC resolver from config: %+v %+v", conf, cfg)
			cli, err = clientv3.New(cfg)
			if err != nil {
				log.Errorf("Failed to create gRPC resolver from config: %+v, err: %v", cfg, err)
				break
			}

			r.cli = cli
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

func (r *etcdv3Client) GetSession() *concurrency.Session {
	if r.cli == nil {
		return nil
	}
	s, err := concurrency.NewSession(r.cli, concurrency.WithTTL(ttl))
	if err != nil {
		log.Errorf("GetSession err: %s", err.Error())
		return nil
	}
	return s
}

type Etcdv3Lock struct {
	mutex *concurrency.Mutex
}

func (r *Etcdv3Lock) Init(lockName string) error {
	s := defaultEtcdv3Client.GetSession()
	if s == nil {
		return errors.New("not ready")
	}
	m := concurrency.NewMutex(s, lockName)
	r.mutex = m
	return nil
}

func (r *Etcdv3Lock) Lock(ctx context.Context) error {
	if err := r.mutex.Lock(ctx); err != nil {
		return err
	}
	return nil
}

func (r *Etcdv3Lock) UnLock(ctx context.Context) error {
	if err := r.mutex.Unlock(ctx); err != nil {
		return err
	}
	return nil
}
