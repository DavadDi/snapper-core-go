package app

import (
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
	"github.com/googollee/go-engine.io"
	"github.com/teambition/jsonrpc"
	"github.com/teambition/snapper-core-go/src/util"
)

// NewClientHandler ...
func NewClientHandler(conn engineio.Conn) *ClientHandler {
	connection := new(ClientHandler)
	connection.conn = conn
	return connection
}

// ClientHandler ...
type ClientHandler struct {
	conn       engineio.Conn
	userid     string
	consumerID string
	source     string
	ioPending  bool
}

func (client *ClientHandler) init() (err error) {
	// auth
	token := client.conn.Request().Header.Get("token")
	jwt, err := jws.ParseJWT([]byte(token))
	err = jwt.Validate([]byte(util.Conf.TokenSecret[0]), crypto.SigningMethodHS256)
	if err != nil {
		log.Println(err.Error())
		return
	}
	payload := jwt.Claims()
	client.userid = payload.Get("userId").(string)
	client.source = payload.Get("source").(string)
	client.consumerID = client.conn.Id()
	consumers.addConsumer(client.consumerID)
	// Bind a consumer to a specified user's room.
	// A user may have one or more consumer's threads.
	consumers.joinRoom("user"+client.userid, client.consumerID)
	consumers.addUserConsumer(client.userid, client.consumerID)
	return
}

func (client *ClientHandler) onMessage(data string) {
	res, _ := jsonrpc.ParseReply(data)
	if res.Type == jsonrpc.RequestType && res.PlayLoad.Error != nil {
		result := res.PlayLoad.Result.(string)
		if result == "OK" {

		}
	}
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
