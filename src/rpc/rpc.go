package rpc

import (
	"net"

	"fmt"

	"log"

	"github.com/teambition/snapper-core-go/src/util"
)

// Start ...
func Start(path string) {
	util.Conf = util.InitConfig(path)
	listener, err := net.Listen("tcp", fmt.Sprint(":", util.Conf.RPCPort))
	if err != nil {
		log.Fatal(err.Error())
	} else {
		for {
			c, err := listener.Accept()
			if err != nil {
				fmt.Print(err.Error())
				continue
			}
			client := NewConnection(c)
			go client.handleConn()
		}
	}
	return
}
