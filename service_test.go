package LiteRPC

import (
	"reflect"
	"testing"
)

type foo struct {
}

func (f *foo) Double(arg int, reply *int) error {
	*reply = arg*2
	return nil
}

func TestService(t *testing.T) {
	s := NewService(&foo{})
	m := s.getMethod("Double")
	argv := m.newArgv()
	replyv := m.newReplyv()
	argv.Set(reflect.ValueOf(5))
	err := s.call(m, argv, replyv)
	if err != nil {
		t.Fatalf("service calling error %s", err.Error())
	}
	res := *replyv.Interface().(*int)
	if res != 10 {
		t.Fatalf("expect %d get %d", 10, res)
	}
}
