package rpc

import (
	"errors"
	"net"

	"fmt"

	"log"

	"github.com/teambition/snapper-core-go/src/util"
)

var (
	errorUnauthorized = errors.New("{\"name\":\"Unauthorized\"}")
)

// RPC ...
type RPC struct {
	listener net.Listener
}

// Start ...
func (rpc *RPC) Start(path string) {
	if rpc.listener != nil {
		log.Fatal("RPC instance already exists")
	}
	util.Conf = util.InitConfig(path)
	var err error
	rpc.listener, err = net.Listen("tcp", fmt.Sprint(":", util.Conf.RPCPort))
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	for {
		c, err := rpc.listener.Accept()
		if err != nil {
			fmt.Print(err.Error())
			break
		}
		client := NewConnection(c)
		go client.handleConn()
	}
	return
}

// Close ...
func (rpc *RPC) Close() (err error) {
	if rpc.listener == nil {
		return errors.New("No running rpc")
	}
	err = rpc.listener.Close()
	rpc.listener = nil
	return
}
