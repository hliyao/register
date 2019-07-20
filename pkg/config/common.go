package config

import (
	beegoConfig "github.com/astaxie/beego/config"
)

// Configer type alias
type Configer = beegoConfig.Configer

var (
	NewConfig     = beegoConfig.NewConfig
	NewConfigData = beegoConfig.NewConfigData
)

type ConfigMap map[string]interface{}

func (m ConfigMap) String(key string, def ...string) string {
	if v, ok := m[key]; ok {
		if vv, ok := v.(string); ok {
			return vv
		}
	}
	if len(def) > 0 {
		return def[0]
	}
	return ""
}

func (m ConfigMap) Float(key string, def ...float64) float64 {
	if v, ok := m[key]; ok {
		if vv, ok := v.(float64); ok {
			return vv
		}
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0.0
}

func (m ConfigMap) Int(key string, def int) int {
	return int(m.Float(key, float64(def)))
}

func (m ConfigMap) Int64(key string, def int64) int64 {
	return int64(m.Float(key, float64(def)))
}
