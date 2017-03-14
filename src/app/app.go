package app

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"encoding/json"

	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
	"github.com/googollee/go-engine.io"
	"github.com/teambition/jsonrpc"
	"github.com/teambition/snapper-core-go/src/util"
)

var consumers = NewConsumers()

// APP ...
type APP struct {
}

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
		for {
			conn, _ := server.Accept()
			handler := NewHandler(conn)
			err = handler.auth()
			if err == nil {
				consumers.Add("", handler)
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

// NewHandler ...
func NewHandler(conn engineio.Conn) *ClientHandler {
	connection := new(ClientHandler)
	connection.conn = conn
	//connection.p = &producer{client: getClient()}
	return connection
}

// ClientHandler ...
type ClientHandler struct {
	conn engineio.Conn
}

func (client *ClientHandler) auth() (err error) {
	// log.Println("connected:", client.conn.Id())
	// defer func() {
	// 	client.conn.Close()
	// 	log.Println("disconnected:", client.conn.Id())
	// }()
	token := client.conn.Request().Header.Get("token")
	jwt, err := jws.ParseJWT([]byte(token))
	err = jwt.Validate([]byte(util.Conf.TokenSecret[0]), crypto.SigningMethodHS256)
	if err != nil {
		log.Println(err.Error())
	}
	return
}
func (client *ClientHandler) init() {
	t, r, err := client.conn.NextReader()
	if err != nil {
		return
	}
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return
	}
	r.Close()
	if t == engineio.MessageText {
		client.onMessage(string(b))
	} else {
		log.Println(t, hex.EncodeToString(b))
	}
	w, err := client.conn.NextWriter(t)
	if err != nil {
		return
	}
	w.Write([]byte("pong"))
	w.Close()
}
func (client *ClientHandler) onMessage(data string) interface{} {
	res, _ := jsonrpc.Parse(data)
	if res.Type == jsonrpc.RequestType {
		res, _ := jsonrpc.Success(res.PlayLoad.ID, res.PlayLoad.Params)
		data, _ := json.Marshal(res)
		return data
	}
	return ""
}

func (client *ClientHandler) sendMessage(datas []string) {
	w, err := client.conn.NextWriter(engineio.MessageText)
	if err != nil {
		return
	}
	res, _ := jsonrpc.Request2("publish", datas)
	data, _ := json.Marshal(res)
	w.Write(data)
	w.Close()

	t, r, err := client.conn.NextReader()
	if err != nil {
		return
	}
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return
	}
	r.Close()
	if t == engineio.MessageText {
		client.onMessage(string(b))
	} else {
		log.Println(t, hex.EncodeToString(b))
	}
}

// Close ...
func (client *ClientHandler) Close() {
	client.conn.Close()
}
