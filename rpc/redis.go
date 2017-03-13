package rpc

import redis "gopkg.in/redis.v5"

var (
	// PREFIX ...
	PREFIX                   = "SNP"
	redisaddr                = "192.168.0.21:6379"
	maxMessageQueueLen int64 = 1024
	defaultRommExp           = 3600 * 24 * 1.5
)

func getClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: redisaddr,
	})
}

// Key for consumer's message queue. It is a List
func genRoomKey(room string) string {
	return PREFIX + ":H:" + room
}

// Key for a room. It is a Hash
func genQueueKey(consumerID string) string {
	return PREFIX + ":L:" + consumerID
}

// Key for a user's state. It is a Set
func genUserStateKey(userID string) string {
	return PREFIX + ":U:" + userID
}
