package codec

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"log"
)

type Header struct {
	ServiceMethod string // format "Service.Method"
	Seq           uint64 // sequence number chosen by client
	Error         string
}

type Codec interface {
	io.Closer
	ReadHeader(*Header) error
	ReadBody(interface{}) error
	Write(*Header, interface{}) error
}

// supporting codec
type Type uint8
type NewCodecFunc func(io.ReadWriteCloser)Codec
const (
	GobCodec Type = iota
	JsonCodec
)
var NewCodecFuncMap map[Type]NewCodecFunc

// number indicating a rpc request
const magicNum uint32 = 0x3bef5cdd

func init()  {
	NewCodecFuncMap = make(map[Type]NewCodecFunc)
	NewCodecFuncMap[GobCodec] = NewGobCodec
}

// option is encoded in bytes slice
// parse option to get the selected codec
func ParseOption(opt []byte) (NewCodecFunc ,error) {
	if len(opt) != 5 {
		log.Println("codec error: error option len")
		return nil, errors.New("codec error: error option len")
	}
	optNum := binary.BigEndian.Uint32(opt[0:4])
	if optNum != magicNum {
		log.Println("codec error: magic number mismatch")
		return nil, errors.New("codec error: magic number mismatch")
	}
	newCodecFunc, ok := NewCodecFuncMap[Type(opt[4])]
	if !ok {
		log.Println("codec error: codec type not supported")
		return nil, errors.New("codec error: codec type not supported")
	}
	return newCodecFunc, nil
}

func GetOption(t Type) []byte {
	opt := make([]byte, 0)
	buf := bytes.NewBuffer(opt)
	_ = binary.Write(buf, binary.BigEndian, magicNum)
	_ = binary.Write(buf, binary.BigEndian, t)
	return buf.Bytes()
}