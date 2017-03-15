package app

import "sync"

// NewClientManager ....
func NewClientManager() *ClientManager {
	return &ClientManager{clients: make(map[string]*ClientHandler)}
}

// ClientManager ...
type ClientManager struct {
	lock    sync.Mutex
	clients map[string]*ClientHandler
}

// Add ...
func (c *ClientManager) Add(key string, client *ClientHandler) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.clients[key] = client

}

// Get ...
func (c *ClientManager) Get(key string) *ClientHandler {
	c.lock.Lock()
	defer c.lock.Unlock()
	ret, ok := c.clients[key]
	if !ok {
		return nil
	}
	return ret
}

// Del ...
func (c *ClientManager) Del(key string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	delete(c.clients, key)
}
