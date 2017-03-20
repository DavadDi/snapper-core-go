package app

import (
	"log"
	"strings"

	"time"

	"github.com/teambition/snapper-core-go/src/service"
	"github.com/teambition/snapper-core-go/src/util"
	redis "gopkg.in/redis.v5"
)

// NewConsumer ...
func NewConsumer() *Consumer {
	consumer := new(Consumer)
	consumer.client = service.GetClient()

	subscribeClient := service.GetClient()
	var err error
	consumer.pubsub, err = subscribeClient.Subscribe(service.GenChannelName())
	if err != nil {
		log.Print(err.Error())
	}
	go consumer.onMessage()
	return consumer
}

// Consumer ...
type Consumer struct {
	client *redis.Client
	pubsub *redis.PubSub
}

// Add consumer queue to redis via websocket.
func (c *Consumer) addConsumer(consumerID string) {
	queueKey := service.GenQueueKey(consumerID)
	str, _ := c.client.LIndex(queueKey, 0).Result()
	if str == "" {
		// Initialize message queue if it does not exist,
		c.client.RPush(queueKey, "1")
	}
	c.client.Expire(queueKey, util.Conf.RedisQueueExpires)
}
func (c *Consumer) updateConsumer(userID, consumerID string) {
	queueKey := service.GenQueueKey(consumerID)
	c.addUserConsumer(userID, consumerID)
	c.client.Expire(queueKey, util.Conf.RedisQueueExpires)
}
func (c *Consumer) onMessage() {
	for {
		msgi, err := c.pubsub.ReceiveTimeout(time.Hour)
		if err != nil {
			log.Print(err.Error())
			break
		}
		switch msg := msgi.(type) {
		case *redis.Message:
			if service.GenChannelName() != msg.Channel {
				return
			}
			consumerIds := strings.Split(msg.Payload, ",")
			for _, id := range consumerIds {
				c.pullMessage(id)
			}
		}
	}
}
func (c *Consumer) pullMessage(consumerID string) {
	client := clientManager.Get(consumerID)
	if client == nil {
		log.Print("no suitable consumer")
		return
	}
	go func() {
		defer func() {
			r := recover()
			if r != nil {
				log.Println(r)
			}
			clientManager.ReleaseIO(consumerID)
		}()
		queueKey := service.GenQueueKey(consumerID)
		// Pull at most 20 messages at a time.
		// A placeholder message is at index 0 (`'1'` or last unread message).
		// Because empty list will be removed automatically.
		msgs, err := c.client.LRange(queueKey, 1, service.DefaultMessagesToPull).Result()
		if err != nil {
			log.Print(err.Error())
		}
		if len(msgs) < 1 {
			return
		}
		client.sendMessage(msgs)
		msglength := int64(len(msgs))
		c.client.LTrim(queueKey, msglength, -1)
		stats.IncrConsumerMessages(msglength)
	}()
}
func (c *Consumer) addUserConsumer(userID, consumerID string) {
	userkey := service.GenUserStateKey(userID)
	c.client.SAdd(userkey, consumerID).Result()
	c.client.Expire(userkey, service.DefaultRoomExp).Result()
	// clean stale consumerId
	c.checkUserConsumers(userID, consumerID)
}

func (c *Consumer) checkUserConsumers(userID, consumerID string) {
	userkey := service.GenUserStateKey(userID)
	consumerIds, err := c.client.SMembers(userkey).Result()
	if err != nil {
		return
	}
	for _, id := range consumerIds {
		if id == consumerID {
			continue
		}
		queueKey := service.GenQueueKey(id)
		count, _ := c.client.LLen(queueKey).Result()
		// count is 0 means consumer not exists.
		// count is great than 1 means consumer is not online, so some messages are not consumed.
		// but sometimes mistake, such as messages are not consumed in time.
		// we rescue it in updateConsumer(socket heartbeat).
		if count != 1 {
			c.removeUserConsumer(userID, id)
		}
	}
}
func (c *Consumer) removeUserConsumer(userID, consumerID string) {
	userkey := service.GenUserStateKey(userID)
	c.client.SRem(userkey, consumerID).Result()
}

func (c *Consumer) weakenConsumer(consumerID string) {
	queueKey := service.GenQueueKey(consumerID)
	c.client.Expire(queueKey, service.DefaultMessageQueueExp).Result()
}

// Add a consumer to a specified room via rpc.
func (c *Consumer) joinRoom(room, consumerID string) (bool, error) {
	roomkey := service.GenRoomKey(room)
	b, err := c.client.HSet(roomkey, consumerID, "1").Result()
	c.client.Expire(roomkey, service.DefaultRoomExp).Result()
	return b, err
}
