package LiteRPC

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

type registry struct {
	mu      sync.Mutex
	servers map[string]*ServerInfo
}

type ServerInfo struct {
	Addr  string
	start time.Time
}

func NewRegistry() *registry {
	r := &registry{
		servers: make(map[string]*ServerInfo),
	}
	http.Handle("/LiteRPC", r)
	return r
}

func (r *registry) getAliveServers() []string {
	servers := make([]string, len(r.servers))
	i := 0
	for k := range r.servers {
		servers[i] = k
		i += 1
	}
	fmt.Println(servers)
	return servers
}

func (r *registry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mu.Lock()
	defer r.mu.Unlock()
	switch req.Method {
	// get registered servers
	case http.MethodGet:
		w.Header().Set("rpc-servers", strings.Join(r.getAliveServers(), ","))
	// register server
	case http.MethodPost:
		addr := req.Header.Get("rpc-server-addr")
		serv := &ServerInfo{
			Addr: addr,
		}
		r.servers[addr] = serv
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
