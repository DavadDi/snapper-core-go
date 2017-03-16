package app

import (
	"encoding/hex"
	"errors"
	"net"
	"strconv"

	"crypto/md5"

	"os"

	"io"

	"github.com/teambition/snapper-core-go/src/service"
	"github.com/teambition/snapper-core-go/src/util"
	redis "gopkg.in/redis.v5"
)

// NewStats ...
func NewStats() *Stats {
	stats := &Stats{
		statsKey:  util.Conf.RedisPrefix + ":STATS",
		roomKey:   util.Conf.RedisPrefix + ":STATS:ROOM",
		serverKey: util.Conf.RedisPrefix + ":STATS:SERVERS",
		client:    service.GetClient(),
	}
	ip, _ := externalIP()
	stats.network = ip + ":" + strconv.Itoa(os.Getpid())
	h := md5.New()
	io.WriteString(h, stats.network)

	stats.serverID = hex.EncodeToString((h.Sum(nil)))
	return stats
}

// Stats ...
type Stats struct {
	client    *redis.Client
	statsKey  string //Hash
	roomKey   string //HyperLogLog
	serverKey string //Hash
	serverID  string
	network   string
}

// Os ...
func (s *Stats) Os() map[string]interface{} {

	result := make(map[string]interface{})
	result["net"] = s.network
	result["serverId"] = s.serverID
	return result
}

// IncrProducerMessages ...
func (s *Stats) IncrProducerMessages(count int64) {
	s.client.HIncrBy(s.statsKey, "producerMessages", count).Result()
}

// IncrConsumerMessages ...
func (s *Stats) IncrConsumerMessages(count int64) {
	s.client.HIncrBy(s.statsKey, "consumerMessages", count).Result()

}

// IncrConsumers ...
func (s *Stats) IncrConsumers(count int64) {
	s.client.HIncrBy(s.statsKey, "consumers", count).Result()

}

// AddRoomsHyperlog  ...
func (s *Stats) AddRoomsHyperlog(roomID string) {
	s.client.PFAdd(s.roomKey, roomID)
}

// SetConsumersStats ...
func (s *Stats) SetConsumersStats(consumers int) {
	s.client.HSet(s.serverKey, s.serverID, string(consumers)).Result()
}

// ClientsStats ...
func (s *Stats) ClientsStats() map[string]interface{} {
	count, _ := s.client.PFCount(s.roomKey).Result()
	stats, _ := s.client.HGetAll(s.statsKey).Result()
	stats["rooms"] = strconv.FormatInt(count, 10)

	servers, _ := s.client.HGetAll(s.serverKey).Result()
	total := 0
	for _, val := range servers {
		num, _ := strconv.Atoi(val)
		total += num
	}

	result := make(map[string]interface{})
	result["total"] = stats
	result["online"] = total
	result["current"] = servers
	return result
}
func externalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}
