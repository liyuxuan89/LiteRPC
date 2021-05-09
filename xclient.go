package LiteRPC

import (
	"LiteRPC/codec"
	"context"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"
)

type SelectMode int

type xClient struct {
	mode    SelectMode
	mu      sync.Mutex
	index   int // index for round robin
	r       *rand.Rand
	addrs   []string // available servers ( get from registry)
	clients map[string]*client
}

type ServerInfo struct {
	Addr string
	Co   codec.Type
}

const (
	RandomSelect SelectMode = iota
	RoundRobinSelect
)

func NewXClient(s SelectMode) *xClient {
	c := &xClient{
		r:       rand.New(rand.NewSource(time.Now().UnixNano())),
		mode:    s,
		clients: make(map[string]*client),
	}
	c.index = c.r.Intn(math.MaxInt32 - 1)
	return c
}

func (xc *xClient) Close() error {
	xc.mu.Lock()
	defer xc.mu.Unlock()
	for k, cli := range xc.clients {
		_ = cli.Close()
		delete(xc.clients, k)
	}
	return nil
}

func (xc *xClient) Dial(addr string, typ codec.Type) (err error) {
	cli, ok := xc.clients[addr]
	if ok && cli.available {
		return nil
	}
	// make sure old connection is closed
	if ok && !cli.available {
		_ = cli.Close()
	}
	// create new client
	cli = NewClient()
	err = cli.Dial(addr, typ)
	if err != nil {
		log.Println("rpc xclient error:", err)
		return err
	}
	xc.clients[addr] = cli
	return nil
}

func (xc *xClient) DialServers(servers []*ServerInfo) (n int, err error) {
	for _, s := range servers {
		err = xc.Dial(s.Addr, s.Co)
		if err == nil {
			n += 1
			xc.addrs = append(xc.addrs, s.Addr)
		}
	}
	return
}

func (xc *xClient) Call(ctx context.Context, serviceMethod string, argv, replyv interface{}) error {
	var idx int
	switch xc.mode {
	case RandomSelect:
		idx = xc.r.Intn(math.MaxInt32-1) % len(xc.addrs)
	case RoundRobinSelect:
		idx = xc.index % len(xc.addrs)
		xc.index += 1
	}
	addr := xc.addrs[idx]
	cli := xc.clients[addr]
	return cli.Call(ctx, serviceMethod, argv, replyv)
}
