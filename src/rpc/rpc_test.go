package rpc

import (
	"testing"

	"time"

	"fmt"

	"github.com/stretchr/testify/assert"
	"github.com/teambition/snapper-core-go/src/util"
	"github.com/teambition/snapper-producer-go"
)

var (
	clientoptions = &snapperproducer.Options{
		SecretKeys: []string{"Usdsiwcs78Ymhpewlk"},
		ExpiresIn:  604800,
		Address:    "127.0.0.1:7700",
		ProducerID: "teambition",
	}
	endsignal chan bool
	userroom  = "user58b7ab90b7162c96120cba9b"
)

func init() {
	endsignal = make(chan bool)
	go func() {

		rpc := new(RPC)
		rpc.Start("../../config/default.json")
		<-endsignal
	}()
	time.Sleep(time.Millisecond * 50)
}
func TestRPC(t *testing.T) {
	t.Run("RPC with receive notifation that should be", func(t *testing.T) {
		// producer sdk
		clientoptions.Address = fmt.Sprint("127.0.0.1:", util.Conf.RPCPort)
		producer, _ := snapperproducer.New(clientoptions)

		sss := "{\"e\":\":change:user/58b7ab90b7162c96120cba9b\",\"d\":{\"ated\":0,\"normal\":14,\"later\":0,\"private\":1,\"badge\":10,\"hasNormal\":true,\"hasAted\":false,\"hasLater\":false,\"hasPrivate\":true}}"
		producer.SendMessage(userroom, sss)
		time.Sleep(time.Millisecond * 50)
	})
	t.Run("RPC with receive request that should be", func(t *testing.T) {
		assert := assert.New(t)

		sss := "{\"e\":\":change:user/58b7ab90b7162c96120cba9b\",\"d\":{\"ated\":0,\"normal\":11,\"later\":0,\"private\":2,\"badge\":10,\"hasNormal\":true,\"hasAted\":false,\"hasLater\":false,\"hasPrivate\":true}}"
		msg := [][]string{[]string{userroom, sss}}

		clientoptions.Address = fmt.Sprint("127.0.0.1:", util.Conf.RPCPort)
		producer, _ := snapperproducer.New(clientoptions)
		result, err := producer.Request("publish", msg)

		assert.Contains(result, "\"result\":1")
		assert.Nil(err)

		msg = [][]string{[]string{userroom, sss}, []string{userroom, sss}}
		result, err = producer.Request("publish", msg)

		assert.Contains(result, "\"result\":2")
		assert.Nil(err)
	})
	t.Run("RPC Close the rpc service", func(t *testing.T) {
		close(endsignal)
	})
}
