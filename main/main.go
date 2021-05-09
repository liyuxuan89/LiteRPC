package main

import (
	"LiteRPC"
	"LiteRPC/codec"
	"context"
	"fmt"
	"log"
	"net"
	"time"
)

type Foo struct{}

type Arg struct {
	Num1       int
	Num2       int
	HandleTime float32
}

func (f *Foo) Double(arg Arg, reply *int) error {
	*reply = arg.Num1 + arg.Num2
	time.Sleep(time.Second * time.Duration(arg.HandleTime))
	return nil
}

func startServer(addr chan<- string) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Println("network error", err)
	}
	log.Println("server runs on", l.Addr().String())
	server := LiteRPC.NewServer(time.Second)
	_ = server.Register(&Foo{})
	addr <- l.Addr().String()
	server.Accept(l)
}

func main() {
	addr1 := make(chan string)
	addr2 := make(chan string)
	go startServer(addr1)
	go startServer(addr2)
	servers := make([]*LiteRPC.ServerInfo, 2)
	servers[0] = &LiteRPC.ServerInfo{
		Addr: <-addr1,
		Co:   codec.GobCodec,
	}
	servers[1] = &LiteRPC.ServerInfo{
		Addr: <-addr2,
		Co:   codec.GobCodec,
	}
	cli := LiteRPC.NewXClient(LiteRPC.RoundRobinSelect)
	n, err := cli.DialServers(servers)
	fmt.Println("servers num", n)
	//cli := LiteRPC.NewClient()
	//err := cli.Dial(<-addr, codec.GobCodec)
	//if err != nil {
	//	log.Println("Dial error", err)
	//}
	//defer func() {
	//	_ = cli.Close()
	//}()
	var ret int
	arg := &Arg{
		Num1: 10,
		Num2: 20,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	for i := 0; i < 5; i++ {
		arg.Num1 = i
		arg.Num2 = i * 2
		arg.HandleTime = 0.5
		err = cli.Call(ctx, "Foo.Double", arg, &ret)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		fmt.Println(ret)
	}
	for i := 0; i < 5; i++ {
		arg.Num1 = i
		arg.Num2 = i * 2
		arg.HandleTime = 2
		err = cli.Call(ctx, "Foo.Double", arg, &ret)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		fmt.Println(ret)
	}
	cancel()
	time.Sleep(time.Second * 2)
}
