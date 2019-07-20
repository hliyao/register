package config

import (
	"os"
	"os/user"
	"path"
)

const (
	ClientConfigPathEnvName = "CLIENT_CONFIG_PATH"
)

// Default client config path is $HOME$/etc/client
func DefaultClientConfigPath() (configPath string) {
	user, err := user.Current()
	if nil == err {
		configPath = path.Join(user.HomeDir, "etc/client/")
	} else {
		configPath = "/etc/client"
	}
	return
}

func ClientConfigPath() string {
	clientConfigPath := os.Getenv(ClientConfigPathEnvName)
	if len(clientConfigPath) == 0 {
		clientConfigPath = DefaultClientConfigPath()
	}
	return clientConfigPath
}
