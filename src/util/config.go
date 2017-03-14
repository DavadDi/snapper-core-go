package util

import (
	"encoding/json"
	"io/ioutil"
	"time"
)

// Conf ...
var Conf *Config

// Config ...
type Config struct {
	Port              int           `json:"port"`
	RPCPort           int           `json:"rpcPort"`
	RedisPrefix       string        `json:"redisPrefix"`
	TokenSecret       []string      `json:"tokenSecret"`
	TokenExpires      int           `json:"tokenExpires"`
	RedisQueueExpires time.Duration `json:"redisQueueExpires"`
	Redis             RedisConfig   `json:"redis"`
	ReadWriteTimeout  time.Duration `json:"readWriteTimeout"`
}

// InitConfig ...
func InitConfig(path string) *Config {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err.Error())
	}
	var config Config
	err = json.Unmarshal(b, &config)
	if err != nil {
		panic(err.Error())
	}
	config.ReadWriteTimeout = config.ReadWriteTimeout * time.Second
	config.RedisQueueExpires = config.RedisQueueExpires * time.Second
	return &config
}

// RedisConfig ...
type RedisConfig struct {
	Hosts   []string    `json:"hosts"`
	Options interface{} `json:"options"`
}
