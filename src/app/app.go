package app

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/googollee/go-engine.io"
	"github.com/teambition/snapper-core-go/src/util"
)

var (
	clientManager *ClientManager
	consumers     *Consumer
)

// Start ...
func Start(path string) {

	util.Conf = util.InitConfig(path)
	server, err := engineio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}
	server.SetPingInterval(time.Second * 30)
	server.SetPingTimeout(time.Second * 5)

	go func() {
		clientManager = NewClientManager()
		consumers = NewConsumer()
		for {
			conn, _ := server.Accept()
			handler := NewClientHandler(conn)
			err = handler.init()
			if err == nil {
				clientManager.Add(handler.consumerID, handler)
			} else {
				handler.Close()
			}
		}
	}()

	http.Handle("/engine.io/", server)
	http.HandleFunc("/", versionHandler)

	log.Fatal(http.ListenAndServe(fmt.Sprint(":", util.Conf.Port), nil))
}

func versionHandler(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, "{\"server\":\"snapper-core\",\"version\":\"0.0.1\"}")
}
