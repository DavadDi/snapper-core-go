package app

import (
	"log"
	"strings"

	"time"

	"sync"

	"github.com/teambition/snapper-core-go/src/service"
	"github.com/teambition/snapper-core-go/src/util"
	redis "gopkg.in/redis.v5"
)

// NewConsumers ....
func NewConsumers() *Consumers {
	return &Consumers{clients: make(map[string]*ClientHandler)}
}

// Consumers ...
type Consumers struct {
	lock    sync.Mutex
	clients map[string]*ClientHandler
}

// Add ...
func (c *Consumers) Add(key string, client *ClientHandler) {
	c.lock.Lock()
	c.clients[key] = client
	c.lock.Unlock()
}

// Get ...
func (c *Consumers) Get(key string) *ClientHandler {
	return c.clients[key]
}

// Del ...
func (c *Consumers) Del(key string) {
	delete(c.clients, key)
}

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
		// Initialize message queue if key does not exist,
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
	client := consumers.Get(consumerID)

	client.sendMessage(msgs)
	c.client.LTrim(queueKey, 1, service.DefaultMessagesToPull)
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
