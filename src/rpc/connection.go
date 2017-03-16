package rpc

import (
	"net"

	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
	"github.com/teambition/jsonrpc"
	"github.com/teambition/respgo"
	"github.com/teambition/snapper-core-go/src/service"
	"github.com/teambition/snapper-core-go/src/util"
	"github.com/teambition/socket-pool-go"
)

// NewConnection ...
func NewConnection(conn net.Conn) *Connection {
	connection := new(Connection)
	connection.Socket = &socketpool.Socket{Conn: conn}
	connection.p = &producer{client: service.GetClient()}
	return connection
}

// Connection ...
type Connection struct {
	*socketpool.Socket
	p *producer
}

func (conn *Connection) handleConn() {
	var err error
	defer func() {
		if err != nil {
			conn.Write([]byte(err.Error()), util.Conf.ReadWriteTimeout)
		}
		conn.Close()
	}()
	for {
		var resp interface{}
		_, resp, err = respgo.Parse(conn.Conn, util.Conf.ReadWriteTimeout)
		msg, ok := resp.(string)
		if err != nil || !ok {
			break
		}
		var req *jsonrpc.ClientRequest
		req, err = jsonrpc.Parse(msg)
		if err != nil || req.Type == jsonrpc.Invalid || req.Type == jsonrpc.ErrorType {
			break
		}
		err = conn.handleJSONRPC(req)
		if err != nil {
			err = nil
			break
		}
	}
}
func (conn *Connection) handleJSONRPC(req *jsonrpc.ClientRequest) (err error) {
	var result interface{}
	var errObj *jsonrpc.ErrorObj
	defer func() {
		r := recover()
		if req.Type == jsonrpc.NotificationType {
			return
		}
		var reply string
		if r != nil {
			reply, _ = jsonrpc.Error(req.PlayLoad.ID, jsonrpc.RPCInternalError)
		} else {
			if err != nil {
				reply, _ = jsonrpc.Error(req.PlayLoad.ID, errObj)
			} else {
				reply, _ = jsonrpc.Success(req.PlayLoad.ID, result)
			}
		}
		resp := respgo.EncodeString(reply)
		conn.Write(resp, util.Conf.ReadWriteTimeout)
	}()
	if req.Type == jsonrpc.RequestType {
		switch req.PlayLoad.Method {
		case "auth":
			result, err = conn.auth(req)
		case "subscribe":
			args := req.PlayLoad.Params.([]interface{})
			result, err = conn.p.joinRoom(args[0].(string), args[1].(string))
		case "unsubscribe":
			args := req.PlayLoad.Params.([]interface{})
			result, err = conn.p.leaveRoom(args[0].(string), args[1].(string))
		case "consumers":
			args := req.PlayLoad.Params.([]interface{})
			result, err = conn.p.getUserConsumers(args[0].(string))
		case "publish":
			result, err = conn.publish(req)
		case "echo":
			result = req.PlayLoad.Params
		default:
			errObj = jsonrpc.RPCNotFound
		}
	} else if req.Type == jsonrpc.NotificationType {
		if req.PlayLoad.Method == "publish" {
			result, err = conn.publish(req)
		} else {
			errObj = jsonrpc.RPCNotFound
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
	err = jwt.Validate([]byte(util.Conf.TokenSecret[0]), crypto.SigningMethodHS256)
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
		err = conn.p.broadcastMessage(array[0].(string), array[1].(string))
		if err != nil {
			return
		}
	}
	return
}
