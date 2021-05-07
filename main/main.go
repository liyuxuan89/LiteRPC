package main

import (
	"LiteRPC"
	"LiteRPC/codec"
	"fmt"
	"log"
	"net"
)

type Foo struct {
}

func (f *Foo) Double(arg int, reply *int) error {
	*reply = arg*2
	return nil
}

func startServer(addr chan<- string) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Println("network error", err)
	}
	log.Println("server runs on", l.Addr().String())
	server := LiteRPC.NewServer()
	_ = server.Register(&Foo{})
	addr <- l.Addr().String()
	server.Accept(l)
}

func main() {
	addr := make(chan string)
	go startServer(addr)
	conn, err := net.Dial("tcp", <-addr)
	if err != nil {
		log.Println("Dial error", err)
	}
	defer conn.Close()

	opt := codec.GetOption(codec.GobCodec)
	_, err = conn.Write(opt) // write loop
	newCodecFunc, _ := codec.ParseOption(opt)
	co := newCodecFunc(conn)

	h := &codec.Header{
		Seq: 100,
		ServiceMethod: "Foo.Double",
	}
	b := 10
	_ = co.Write(h, &b)

	co.ReadHeader(h)
	fmt.Println(h)
	co.ReadBody(&b)
	fmt.Println(b)
}
