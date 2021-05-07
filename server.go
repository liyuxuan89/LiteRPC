package LiteRPC

import (
	"LiteRPC/codec"
	"errors"
	"io"
	"log"
	"net"
	"strings"
	"sync"
)

type Server struct {
	serviceMap sync.Map
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Register(recv interface{}) error {
	serve := NewService(recv)
	_, dup := s.serviceMap.LoadOrStore(serve.name,serve)
	if dup {
		log.Println("rpc server: service already loaded: " + serve.name)
		return errors.New("rpc server: service already loaded: " + serve.name)
	}
	return nil
}

func (s *Server) Accept(lis net.Listener)  {
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println("rpc server: accept error", err)
			return
		}
		go s.ServeConn(conn)
	}
}

func (s *Server) ServeConn(conn io.ReadWriteCloser)  {
	// ****************************** read option ******************************
	opt := make([]byte, 5)
	for readBytes := 0; readBytes < 5; {
		n, err := conn.Read(opt[readBytes:])
		if err != nil {
			if err != io.EOF {
				log.Println("rpc server: conn read error", err)
			}
			break
		}
		readBytes += n
	}
	// ****************************** creating corresponding codec ******************************
	newCodecFunc, err := codec.ParseOption(opt)
	if err != nil {
		log.Println("rpc server: parsing option error")
		return
	}
	s.ServeCodec(newCodecFunc(conn))
}

func (s *Server) ServeCodec(c codec.Codec)  {
	header := new(codec.Header)
	var err error
	for {
		// ****************************** read header ******************************
		err = c.ReadHeader(header)
		if err != nil {
			log.Println("rpc server: parsing header error")
			break
		}
		// ****************************** get service method ******************************
		serviceMethod := header.ServiceMethod
		serviceMethodStrings := strings.Split(serviceMethod, ".")
		if len(serviceMethodStrings) != 2 {
			log.Println("rpc server: ill formed service method")
			break
		}
		servei, ok := s.serviceMap.Load(serviceMethodStrings[0])
		if !ok {
			log.Println("rpc server: request service unavailable")
			break
		}
		serve := servei.(*service)
		methodTyp := serve.getMethod(serviceMethodStrings[1])
		if methodTyp == nil {
			log.Println("rpc server: request method unavailable")
			break
		}
		// ****************************** get argv and replyv ******************************
		argv := methodTyp.newArgv()
		replyv := methodTyp.newReplyv()
		body := argv.Addr().Interface()
		err = c.ReadBody(body)
		if err != nil {
			log.Println("rpc server: parsing body error")
			break
		}
		err = serve.call(methodTyp, argv, replyv)
		var replyvi interface{}
		if err != nil {
			log.Println("rpc server: calling error " + err.Error())
			header.Error = err.Error()
			replyvi = nil
		} else {
			replyvi = replyv.Interface()
		}
		// ****************************** send response ******************************
		err = c.Write(header, replyvi)
		if err != nil {
			log.Println("rpc server: write response error")
			break
		}
	}
}
