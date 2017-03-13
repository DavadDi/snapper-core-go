package rpc

import (
	"time"

	"github.com/teambition/snapper-core-go/util"
	redis "gopkg.in/redis.v5"
)

var (
	maxMessageQueueLen int64 = 1024
	defaultRoomExp           = time.Duration(3600*24*1.5) * time.Second
)

func getClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: util.Conf.Redis.Hosts[0],
	})
}

// Key for consumer's message queue. It is a List
func genRoomKey(room string) string {
	return util.Conf.RedisPrefix + ":H:" + room
}

// Key for a room. It is a Hash
func genQueueKey(consumerID string) string {
	return util.Conf.RedisPrefix + ":L:" + consumerID
}

// Key for a user's state. It is a Set
func genUserStateKey(userID string) string {
	return util.Conf.RedisPrefix + ":U:" + userID
}

func genChannelName() string {
	return util.Conf.RedisPrefix + ":message"
}
