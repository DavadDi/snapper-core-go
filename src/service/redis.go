package service

import (
	"time"

	"github.com/teambition/snapper-core-go/src/util"
	redis "gopkg.in/redis.v5"
)

var (
	// DefaultMessageQueueExp ...
	DefaultMessageQueueExp = time.Duration(60*5) * time.Second
	// MaxMessageQueueLen ...
	MaxMessageQueueLen int64 = 1024
	// DefaultRoomExp ...
	DefaultRoomExp = time.Duration(3600*24*1.5) * time.Second
	// DefaultMessagesToPull pull messages from redis queue once
	DefaultMessagesToPull int64 = 50
)

// GetClient ...
func GetClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: util.Conf.Redis.Hosts[0],
	})
}

// GenQueueKey Key for consumer's message queue. It is a List
func GenQueueKey(consumerID string) string {
	return util.Conf.RedisPrefix + ":L:" + consumerID
}

//GenRoomKey Key for a room. It is a Hash
func GenRoomKey(room string) string {
	return util.Conf.RedisPrefix + ":H:" + room
}

// GenUserStateKey Key for a user's state. It is a Set
func GenUserStateKey(userID string) string {
	return util.Conf.RedisPrefix + ":U:" + userID
}

// GenChannelName ...
func GenChannelName() string {
	return util.Conf.RedisPrefix + ":message"
}
