package config

import (
	"encoding/json"
	"io"
	"os"
	"path"

	"github.com/coreos/etcd/clientv3"
)

const etcdClientConf = "etcd_client.json"

func GetDefaultEtcdClientConfig() (cfg clientv3.Config, err error) {
	clientConfigPath := os.Getenv(ClientConfigPathEnvName)
	if len(clientConfigPath) == 0 {
		clientConfigPath = DefaultClientConfigPath()
	}

	var f io.ReadCloser
	f, err = os.Open(path.Join(clientConfigPath, etcdClientConf))
	if err != nil {
		return
	}
	defer f.Close()

	if err = json.NewDecoder(f).Decode(&cfg); err != nil {
		return
	}
	return
}
