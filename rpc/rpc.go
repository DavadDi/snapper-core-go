package rpc

import (
	"errors"
	"net"
	"time"

	"fmt"

	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
	"github.com/teambition/jsonrpc"
	"github.com/teambition/respgo"
	"github.com/teambition/socket-pool-go"
)

var (
	errorUnauthorized = errors.New("{\"name\":\"Unauthorized\"}")
	readWriteTimeout  = 30 * time.Second
)

// StartRPC ...
func StartRPC() (err error) {
	l, err := net.Listen("tcp", ":8888")
	if err != nil {
		fmt.Print(err.Error())
		return
	}
	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Print(err.Error())
			break
		}
		client := NewConnection(c)
		go client.handleConn()
	}
	return
}

// NewConnection ...
func NewConnection(conn net.Conn) *Connection {
	return &Connection{Socket: socketpool.Socket{Conn: conn}}
}

// Connection ...
type Connection struct {
	socketpool.Socket
}

func (conn *Connection) handleConn() {
	defer conn.Close()
	for {
		resp, err := conn.ReadAll(readWriteTimeout)
		if err != nil {
			conn.Write([]byte(err.Error()), readWriteTimeout)
			break
		}
		rpc, err := respgo.DecodeToString(resp)
		if err != nil {
			conn.Write([]byte(err.Error()), readWriteTimeout)
			break
		}
		res, err := jsonrpc.Parse(string(rpc))
		if err != nil || res.Type == jsonrpc.Invalid || res.Type == jsonrpc.ErrorType {
			conn.Write([]byte(err.Error()), readWriteTimeout)
			break
		}
		err = conn.handleJSONRPC(res)
		if err != nil {
			break
		}
	}
}
func (conn *Connection) handleJSONRPC(req *jsonrpc.ClientRequest) (err error) {
	var result interface{}
	defer func() {
		r := recover()
		if req.Type == jsonrpc.NotificationType {
			return
		}
		var reply string
		if r != nil {
			reply, _ = jsonrpc.Error(req.PlayLoad.ID, jsonrpc.CreateError(6789, "Internal error"))
		} else {
			if err != nil {
				reply, _ = jsonrpc.Error(req.PlayLoad.ID, jsonrpc.CreateError(6789, err.Error()))
			} else {
				reply, _ = jsonrpc.Success(req.PlayLoad.ID, result)
			}
		}
		resp := respgo.EncodeString(reply)
		conn.Write(resp, readWriteTimeout)
		fmt.Print("handleJSONRPC:" + reply)
	}()
	if req.Type == jsonrpc.RequestType {
		switch req.PlayLoad.Method {
		case "auth":
			result, err = conn.auth(req)
		case "subscribe":
			args := req.PlayLoad.Params.([]interface{})
			joinRoom(args[0].(string), args[1].(string))
		case "unsubscribe":
			args := req.PlayLoad.Params.([]interface{})
			leaveRoom(args[0].(string), args[1].(string))
		case "consumers":
			args := req.PlayLoad.Params.([]interface{})
			getUserConsumers(args[0].(string))
		case "publish":
			result, err = conn.publish(req)
		}
	} else if req.Type == jsonrpc.NotificationType {
		if req.PlayLoad.Method == "publish" {
			result, err = conn.publish(req)
		}
	}
	return
}
func (conn *Connection) auth(req *jsonrpc.ClientRequest) (result interface{}, err error) {
	args := req.PlayLoad.Params.([]interface{})
	if len(args) < 1 {
		return nil, errorUnauthorized
	}
	token, ok := args[0].(string)
	if !ok {
		return nil, errorUnauthorized
	}
	jwt, err := jws.ParseJWT([]byte(token))
	err = jwt.Validate([]byte("Usdsiwcs78Ymhpewlk"), crypto.SigningMethodHS256)
	if err == nil {
		result = "OK"
	} else {
		err = errorUnauthorized
	}
	return
}

// data.params: [
//   [room1, message1],
//   [room2, message2]
//   ...
// ]
func (conn *Connection) publish(req *jsonrpc.ClientRequest) (count int, err error) {
	results := req.PlayLoad.Params.([]interface{})
	for _, args := range results {
		array := args.([]interface{})
		count++
		err = broadcastMessage(array[0].(string), array[1].(string))
		if err != nil {
			return
		}
	}
	return
}
