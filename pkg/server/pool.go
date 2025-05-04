package server

import (
	"k8s.io/client-go/kubernetes"
	"sync"
)

type ClientManager struct {
	clientPool map[string]*sync.Pool
	m          sync.Mutex
}

func NewClientManager(clientMap map[string]*kubernetes.Clientset) *ClientManager {
	pool := make(map[string]*sync.Pool)
	for cluster, client := range clientMap {
		pool[cluster] = &sync.Pool{
			New: func() interface{} {
				return client
			},
		}
	}
	return &ClientManager{
		clientPool: pool,
	}
}

func (c *ClientManager) Get(cluster string) *kubernetes.Clientset {
	pool, ok := c.clientPool[cluster]
	if !ok {
		return nil
	}
	return pool.Get().(*kubernetes.Clientset)
}

func (c *ClientManager) Put(cluster string, client *kubernetes.Clientset) {
	pool, exist := c.clientPool[cluster]
	if !exist {
		return
	}
	pool.Put(client)
}

func (c *ClientManager) set(cluster string, client *kubernetes.Clientset) {
	if client == nil {
		return
	}
	c.m.Lock()
	defer c.m.Unlock()
	c.clientPool[cluster] = &sync.Pool{
		New: func() interface{} {
			return client
		},
	}
	return
}
