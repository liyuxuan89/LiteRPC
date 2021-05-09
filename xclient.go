package LiteRPC

import (
	"LiteRPC/codec"
	"context"
	"errors"
	"log"
	"math"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

type SelectMode int

type xClient struct {
	addrRegistry string
	timeout      time.Duration
	mode         SelectMode
	mu           sync.Mutex
	isClose      bool
	index        int // index for round robin
	r            *rand.Rand
	addrs        []string // available servers ( get from registry)
	clients      map[string]*client
}

const (
	RandomSelect SelectMode = iota
	RoundRobinSelect
)

func NewXClient(s SelectMode, regAddr string) *xClient {
	c := &xClient{
		addrRegistry: regAddr,
		timeout:      60 * time.Second,
		r:            rand.New(rand.NewSource(time.Now().UnixNano())),
		mode:         s,
		isClose:      false,
		clients:      make(map[string]*client),
	}
	c.index = c.r.Intn(math.MaxInt32 - 1)
	go c.getServers()
	return c
}

func (xc *xClient) Close() error {
	xc.mu.Lock()
	defer xc.mu.Unlock()
	for k, cli := range xc.clients {
		_ = cli.Close()
		delete(xc.clients, k)
	}
	xc.isClose = true
	return nil
}

func (xc *xClient) Dial(addr string, typ codec.Type) (err error) {
	xc.mu.Lock()
	defer xc.mu.Unlock()
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

func (xc *xClient) DialServers(servers []string, co codec.Type) (n int, err error) {
	for _, s := range servers {
		err = xc.Dial(s, co)
		if err == nil {
			n += 1
			xc.addrs = append(xc.addrs, s)
		}
	}
	return
}

func (xc *xClient) getServers() {
	for {
		resp, err := http.Get(xc.addrRegistry)
		if err == nil {
			serversString := resp.Header.Get("rpc-servers")
			servers := strings.Split(serversString, ",")

			_, _ = xc.DialServers(servers, codec.GobCodec)
		}
		if xc.isClose {
			break
		}
		select {
		case <-time.After(xc.timeout):
			continue
		}
	}

}

func (xc *xClient) Call(ctx context.Context, serviceMethod string, argv, replyv interface{}) error {
	var idx int
	if len(xc.addrs) == 0 {
		return errors.New("rpc xclient error: not server available")
	}
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
