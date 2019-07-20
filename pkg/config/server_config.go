package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"

	"log"
)

// ServerConfig ...
type ServerConfig struct {
	Configer Configer

	Listen      string `json:"listen" yaml:"listen"`
	DebugListen string `json:"debug_listen" yaml:"debug_listen"`
	RestListen  string `json:"rest_listen" yaml:"rest_listen"`
	Name        string `json:"name" yaml:"name"`     // Name of this server, such as `exchange_logic`
	Domain      string `json:"domain" yaml:"domain"` // Domain in which this server is distribute, such as `52tt.local`
	Addr        string `json:"addr" yaml:"addr"`     // Addr of this server, by default, it's the same as `Listen`
	LogLevel    string `json:"log_level" yaml:"log_level"`

	HadManualSetListen bool
}

// EmptyServerConfig returns an empty config
func EmptyServerConfig() *ServerConfig {
	sc := &ServerConfig{
		Listen:   ":",
		LogLevel: "info",
	}
	return sc
}

func (sc *ServerConfig) initWithConfiger(conf Configer) {
	const (
		defaultLogLevel = "info"
	)

	sc.Configer = conf

	sc.Listen = conf.String("server::listen")
	sc.DebugListen = conf.String("server::debug_listen")
	sc.RestListen = conf.String("server::rest_listen")
	sc.LogLevel = conf.DefaultString("server::log_level", defaultLogLevel)
	sc.Name = conf.String("server::name")
	sc.Domain = conf.String("server::domain")
	sc.Addr = conf.String("server::addr")
}

// InitWithPath ...
func (sc *ServerConfig) InitWithPath(configType string, filePath string) error {
	/**
	  *	An Example of some_server.json
	  *  @code
	  # some_server.json
	  {
	    "server": {
		  "listen": ":18151",
		  "log_level": "debug",
		  "name": "pushnotification_v2",
		  "domain": "52tt.local",
		  "addr": "192.168.80.64:18151"
	    }
	  }
	  * @endcode
	  *
	*/
	switch configType {
	case "ini", "json":
		// TODO(wuwei): remove dep to beego
		conf, err := NewConfig(configType, filePath)
		if err != nil {
			log.Fatalln("Failed to init ServerConfig from file:", filePath)
			return err
		}

		sc.initWithConfiger(conf)
		return nil
	case "yaml":
		f, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer f.Close()

		var cfg = struct {
			Server *ServerConfig `yaml:"server"`
		}{
			Server: sc,
		}
		return yaml.NewDecoder(f).Decode(&cfg)
	}

	return errors.New("unknown config type (only ini/json/yaml are supported)")
}

func (sc *ServerConfig) GetLogLevel() string{
	//TODO
	return ""
}

func (sc *ServerConfig) FullName() string {
	if len(sc.Name) == 0 {
		return ""
	}
	if len(sc.Domain) == 0 {
		return sc.Name
	}
	return fmt.Sprintf("%s.%s", sc.Name, sc.Domain)
}
