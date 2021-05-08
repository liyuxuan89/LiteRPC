package LiteRPC

import (
	"LiteRPC/codec"
	"context"
	"errors"
	"log"
	"net"
	"sync"
)

type call struct {
	seq    uint64
	done   chan struct{} // signal call finish
	err    error         // calling fail
	replyv interface{}   // response data
}

type client struct {
	mu        sync.Mutex
	cc        codec.Codec
	header    *codec.Header // sending is mutually exclusive, all request share the save header
	pending   map[uint64]*call
	available bool
}

func NewClient() *client {
	return &client{
		header:    new(codec.Header),
		pending:   make(map[uint64]*call),
		available: false,
	}
}

func (c *client) Close() error {
	c.available = false
	return c.cc.Close()
}

func (c *client) Dial(addr string, typ codec.Type) error {
	opt := codec.GetOption(typ)
	newCodecFunc, err := codec.ParseOption(opt)
	if err != nil {
		log.Println("rpc client: parsing option error:", err)
		return err
	}
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Println("rpc client: dial error:", err)
		return err
	}
	c.cc = newCodecFunc(conn)
	writeBytes := 0
	for writeBytes < len(opt) {
		n, err := conn.Write(opt)
		if err != nil {
			log.Println("rpc client: write option error:", err)
			return err
		}
		writeBytes += n
	}
	c.available = true
	go c.receive()
	return nil
}

func (c *client) Call(ctx context.Context, serviceMethod string, argv, replyv interface{}) (err error) {
	done := make(chan struct{})
	ca := new(call)
	go func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.header.Seq += 1
		c.header.ServiceMethod = serviceMethod
		c.header.Error = ""
		err = c.cc.Write(c.header, argv)
		if err != nil {
			log.Println("rpc client: write request error:", err)
			return
		}
		ca.seq = c.header.Seq
		ca.replyv = replyv
		ca.done = done
		c.pending[c.header.Seq] = ca
	}()
	select {
	case <-ctx.Done():
		c.removeCall(ca.seq)
		return errors.New("rpc client: calling timeout")
	case <-done:
		if ca.err != nil {
			log.Println("rpc client: getting replyv error:", err)
			return err
		}
	}
	return nil
}

func (c *client) removeCall(seq uint64) *call {
	c.mu.Lock()
	defer c.mu.Unlock()
	ca := c.pending[seq]
	delete(c.pending, seq)
	return ca
}

func (c *client) terminateCalls(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, ca := range c.pending {
		ca.err = err
		close(ca.done)
		delete(c.pending, k)
	}
}

func (c *client) receive() {
	var err error
	for err == nil {
		h := new(codec.Header)
		if err = c.cc.ReadHeader(h); err != nil {
			break
		}
		ca := c.removeCall(h.Seq)
		switch {
		case ca == nil:
			err = errors.New("rpc client: receive seq doesn't exist")
		case h.Error != "":
			err = errors.New("rpc client: " + h.Error)
			ca.err = err
			close(ca.done)
		default:
			err = c.cc.ReadBody(ca.replyv)
			ca.err = err
			close(ca.done)
		}
	}
	c.terminateCalls(err)
	c.available = false
	_ = c.Close()
}
