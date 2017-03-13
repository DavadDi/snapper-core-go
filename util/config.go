package util

import (
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
	RedisQueueExpires int           `json:"redisQueueExpires"`
	Redis             RedisConfig   `json:"redis"`
	ReadWriteTimeout  time.Duration `json:"readWriteTimeout"`
}

// RedisConfig ...
type RedisConfig struct {
	Hosts   []string    `json:"hosts"`
	Options interface{} `json:"options"`
}
