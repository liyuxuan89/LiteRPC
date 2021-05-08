package main

import (
	"LiteRPC"
	"LiteRPC/codec"
	"fmt"
	"log"
	"net"
)

type Foo struct{}

type Arg struct {
	Num1 int
	Num2 int
}

func (f *Foo) Double(arg Arg, reply *int) error {
	*reply = arg.Num1 + arg.Num2
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
	cli := LiteRPC.NewClient()
	err := cli.Dial(<-addr, codec.GobCodec)
	// conn, err := net.Dial("tcp", <-addr)
	if err != nil {
		log.Println("Dial error", err)
	}
	defer cli.Close()
	var ret int
	arg := &Arg{
		Num1: 10,
		Num2: 20,
	}
	for i := 0; i < 5; i++ {
		arg.Num1 = i
		arg.Num2 = i * 2
		_ = cli.Call("Foo.Double", arg, &ret)
		fmt.Println(ret)
	}
}
