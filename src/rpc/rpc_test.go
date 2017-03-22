package rpc

import (
	"testing"

	"time"

	"github.com/stretchr/testify/assert"

	"github.com/teambition/snapper-producer-go/snapper"
)

var (
	clientoptions = &snapper.Options{
		SecretKeys: []string{"Usdsiwcs78Ymhpewlk"},
		ExpiresIn:  604800,
		Address:    "127.0.0.1:7700",
		ProducerID: "teambition",
	}
	endsignal chan bool
	userid    = "58b7ab90b7162c96120cba9b"
	userroom  = "user" + userid
)

func init() {
	endsignal = make(chan bool)
	go func() {
		Start("../../config/default.json")
		<-endsignal
	}()
	time.Sleep(time.Millisecond * 100)
}

func TestRPC(t *testing.T) {
	cases := []string{"{\"e\":\":change:user/58b7ab90b7162c96120cba9b\",\"d\":{\"ated\":0,\"normal\":14,\"later\":0,\"private\":1,\"badge\":10,\"hasNormal\":true,\"hasAted\":false,\"hasLater\":false,\"hasPrivate\":true}}"}
	producer, _ := snapper.New(clientoptions)
	t.Run("RPC with receive notifation that should be", func(t *testing.T) {
		assert := assert.New(t)
		clientoptions = &snapper.Options{
			SecretKeys: []string{"Usdsiwcs78Ymhpewl"},
			ExpiresIn:  604800,
			Address:    "127.0.0.1:7700",
			ProducerID: "teambition",
		}
		_, err := snapper.New(clientoptions)
		assert.Equal(err.Error(), "{\"name\":\"Unauthorized\"}")
	})
	t.Run("RPC with receive notifation that should be", func(t *testing.T) {
		// producer sdk
		producer.SendMessage(userroom, cases[0])
		time.Sleep(time.Millisecond * 50)
	})
	t.Run("RPC with receive message that should be", func(t *testing.T) {
		assert := assert.New(t)
		done := make(chan bool)

		msg := [][]string{[]string{userroom, cases[0]}}
		producer.Request("publish", msg, func(result interface{}, err error) {
			assert.Equal(result, float64(1))
			assert.Nil(err)
			done <- true
		})
		<-done

		msg = [][]string{[]string{userroom, cases[0]}, []string{userroom, cases[0]}}
		producer.Request("publish", msg, func(result interface{}, err error) {
			assert.Equal(result, float64(2))
			assert.Nil(err)
			done <- true
		})
		<-done

		producer.Request("get", msg, func(result interface{}, err error) {
			assert.Equal(result, nil)
			assert.Equal(err.Error(), "Method not found")
			done <- true
		})
		<-done

		producer.Request("consumers", []string{userid}, func(result interface{}, err error) {
			assert.Nil(err)
			assert.Equal(result, map[string]interface{}{"android": float64(0), "ios": float64(0), "length": float64(1), "web": float64(1)})
			done <- true
		})
		<-done

		producer.Request("echo", 123, func(result interface{}, err error) {
			assert.Equal(result, float64(123))
			done <- true
		})
		<-done

		producer.Request("consumers", "1", func(result interface{}, err error) {
			assert.NotNil(err)
			assert.Contains(err.Error(), "interface {} is string, not []interface {}")
			done <- true
		})
		<-done

	})
	t.Run("RPC with JoinRoom and LeaveRoom that should be", func(t *testing.T) {
		assert := assert.New(t)

		done := make(chan bool)
		producer.JoinRoom(userroom, "xxxxxxxxxxxxxxx", func(result interface{}, err error) {
			assert.Equal(result, float64(1))
			assert.Nil(err)
			done <- true
		})
		<-done
		producer.LeaveRoom(userroom, "xxxxxxxxxxxxxxx", func(result interface{}, err error) {
			assert.Equal(result, float64(1))
			assert.Nil(err)
			done <- true
		})
		<-done
		producer.Close()

	})
	t.Run("RPC Close the rpc service", func(t *testing.T) {
		close(endsignal)
	})
}
