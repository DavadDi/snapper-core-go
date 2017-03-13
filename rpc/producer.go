package rpc

import "fmt"
import "time"

var (
	luasha1 = ""
)

// Add a consumer to a specified room via rpc.
func joinRoom(room, consumerID string) bool {
	roomkey := genRoomKey(room)
	client := getClient()
	b, _ := client.HSet(roomkey, consumerID, "1").Result()
	client.Expire(roomkey, time.Duration(defaultRommExp)*time.Second).Result()
	return b
}

// Remove a consumer from a specified room via rpc.
func leaveRoom(room, consumerID string) {
	client := getClient()
	roomkey := genRoomKey(room)
	client.HDel(roomkey, consumerID)
}

// For testing purposes.
func clearRoom(room string) {
	client := getClient()
	roomkey := genRoomKey(room)
	client.Del(roomkey)
}

// Broadcast messages to redis queue
func broadcastMessage(room string, msg string) (err error) {
	roomkey := genRoomKey(room)
	fmt.Print(roomkey)
	client := getClient()
	if luasha1 == "" {
		luasha1, err = client.ScriptLoad(lua).Result()
	}
	fmt.Print(luasha1)
	result, err := client.EvalSha(luasha1, []string{roomkey}).Result()
	if err != nil {
		luasha1, err = client.ScriptLoad(lua).Result()
		if err == nil {
			result, err = client.EvalSha(luasha1, []string{roomkey}).Result()
		} else {
			return
		}
	}
	array := result.([]interface{})
	if len(array) < 1 {
		fmt.Print("not found consumerID")
		return
	}
	consumers := ""
	for _, args := range array {
		consumerID := args.(string)
		consumers = consumers + "," + consumerID
		queueKey := genQueueKey(consumerID)
		res, err := client.RPushX(queueKey, msg).Result()
		if err != nil || res < 1 {
			break
		}
		// Weaken non-exists consumer, it will be removed in next cycle unless it being added again.
		client.HIncrBy(roomkey, consumerID, -1)
		// if queue's length is too large, means that consumer was offline long time,
		// or some exception messages produced. Anyway, it is no need to cache
		if res > maxMessageQueueLen {
			client.LTrim(queueKey, 0, maxMessageQueueLen)
		}
	}
	if len(consumers) > 0 {
		client.Publish(PREFIX+":message", consumers[1:])
	}
	return
}

func getUserConsumers(userID string) {
	// client := getClient()
	// consumerIds, err := client.SMembers(genUserStateKey(userID)).Result()
	// for index, val := range consumerIds {

	// }
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
