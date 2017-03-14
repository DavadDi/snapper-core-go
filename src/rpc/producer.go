package rpc

import (
	"fmt"

	"github.com/teambition/snapper-core-go/src/service"
	"github.com/teambition/snapper-core-go/src/util"

	redis "gopkg.in/redis.v5"
)

type producer struct {
	client  *redis.Client
	luasha1 string
}

// Add a consumer to a specified room via rpc.
func (p *producer) joinRoom(room, consumerID string) (bool, error) {
	roomkey := service.GenRoomKey(room)
	b, err := p.client.HSet(roomkey, consumerID, "1").Result()
	p.client.Expire(roomkey, service.DefaultRoomExp).Result()
	return b, err
}

// Remove a consumer from a specified room via rpc.
func (p *producer) leaveRoom(room, consumerID string) (int64, error) {
	roomkey := service.GenRoomKey(room)
	return p.client.HDel(roomkey, consumerID).Result()
}

// For testing purposes.
func (p *producer) clearRoom(room string) (int64, error) {
	roomkey := service.GenRoomKey(room)
	return p.client.Del(roomkey).Result()
}

// Broadcast messages to redis queue
func (p *producer) broadcastMessage(room string, msg string) (err error) {
	roomkey := service.GenRoomKey(room)
	fmt.Print(roomkey)
	if p.luasha1 == "" {
		p.luasha1, err = p.client.ScriptLoad(lua).Result()
	}
	result, err := p.client.EvalSha(p.luasha1, []string{roomkey}).Result()
	if err != nil {
		p.luasha1, err = p.client.ScriptLoad(lua).Result()
		if err == nil {
			result, err = p.client.EvalSha(p.luasha1, []string{roomkey}).Result()
		} else {
			return
		}
	}
	array := result.([]interface{})
	if len(array) < 1 {
		fmt.Print("not found consumerID")
		return nil
	}
	consumers := ""
	for _, args := range array {
		consumerID := args.(string)
		consumers = consumers + "," + consumerID
		queueKey := service.GenQueueKey(consumerID)
		res, err := p.client.RPushX(queueKey, msg).Result()
		if err != nil || res < 1 {
			break
		}
		// Weaken non-exists consumer, it will be removed in next cycle unless it being added again.
		p.client.HIncrBy(roomkey, consumerID, -1)
		// if queue's length is too large, means that consumer was offline long time,
		// or some exception messages produced. Anyway, it is no need to cache
		if res > service.MaxMessageQueueLen {
			p.client.LTrim(queueKey, 0, service.MaxMessageQueueLen)
		}
	}
	if len(consumers) > 0 {
		p.client.Publish(service.GenChannelName(), consumers[1:])
	}
	return
}

func (p *producer) getUserConsumers(userID string) (ps map[string]int, err error) {
	ps = make(map[string]int)
	consumerIds, err := p.client.SMembers(service.GenUserStateKey(userID)).Result()
	if err != nil {
		return
	}
	ps["length"] = len(consumerIds)
	ps["android"] = 0
	ps["ios"] = 0
	ps["web"] = 0
	for _, val := range consumerIds {
		t := util.IDToSource(val)
		ps[t]++
	}
	return
}

//-- Checkout room' consumers
// -- KEYS[1] room key
const lua string = `
local consumers = redis.call('hgetall', KEYS[1])
local result = {}

if #consumers == 0 then return result end

for index = 1, #consumers, 2 do
  if consumers[index + 1] == '1' then
    result[#result + 1] = consumers[index]
  else
    -- delete stale consumer
    redis.call('hdel', KEYS[1], consumers[index])
  end
end

if #result == 0 then
  redis.call('del', KEYS[1])
else
  -- update room's expire time
  redis.call('expire', KEYS[1], 172800)
end

return result`
