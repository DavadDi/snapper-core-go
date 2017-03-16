package app

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"encoding/json"

	"github.com/googollee/go-engine.io"
	"github.com/teambition/jsonrpc"
	"github.com/teambition/snapper-core-go/src/util"
)

var (
	clientManager *ClientManager
	consumers     *Consumer
	stats         *Stats
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
		stats = NewStats()
		for {
			conn, _ := server.Accept()
			handler := NewClientHandler(conn)
			err = handler.init()
			if err == nil {
				clientManager.Add(handler.consumerID, handler)
			} else {
				log.Println(err.Error())
				w, err := conn.NextWriter(engineio.MessageText)
				if err != nil {
					return
				}
				res, _ := jsonrpc.Request2("publish", "datas")
				data, _ := json.Marshal(res)
				w.Write(data)
				w.Close()
				handler.Close()
			}
		}
	}()

	http.Handle("/websocket/", server)
	http.HandleFunc("/stats/", statsHandler)
	http.HandleFunc("/", versionHandler)

	log.Fatal(http.ListenAndServe(fmt.Sprint(":", util.Conf.Port), nil))
}

func versionHandler(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, "{\"server\":\"snapper-core\",\"version\":\"0.0.1\"}")
}
func statsHandler(rw http.ResponseWriter, req *http.Request) {
	result := make(map[string]interface{})
	result["os"] = stats.Os()
	result["stats"] = stats.ClientsStats()
	out, _ := json.Marshal(result)
	fmt.Fprintf(rw, string(out))
}
