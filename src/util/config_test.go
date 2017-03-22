package util

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	t.Run("RPC with initConfig that should be", func(t *testing.T) {
		assert := assert.New(t)
		config := InitConfig("../../config/default.json")
		assert.Equal(7700, config.RPCPort)
		assert.Equal(time.Duration(30)*time.Second, config.ReadWriteTimeout)

		assert.Panics(func() {
			config = InitConfig("../config/default.json")
		})

		assert.Panics(func() {
			config = InitConfig("../../config/defaulterr.json")
		})
	})
}
