package proxy

import (
	"sync"
)

type RouteTable struct {
	mu sync.RWMutex
	routes map[string]string
}

func NewRouteTable() *RouteTable {
	return &RouteTable{
		routes: make(map[string]string),
	}
}

func (rt *RouteTable) Register(appName, targetURL string) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	rt.routes[appName] = targetURL
}

func (rt *RouteTable) Lookup(appName string) (string, bool) {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	target, ok := rt.routes[appName]
	return target, ok
}

func (rt *RouteTable) List() map[string]string {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	snapshot := make(map[string]string, len(rt.routes))
	for k,v := range rt.routes {
		snapshot[k] = v
	}
	return snapshot
}
