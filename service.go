package LiteRPC

import (
	"go/ast"
	"reflect"
)

type methodType struct {
	method reflect.Method
	ArgType reflect.Type
	ReplyType reflect.Type
}

type service struct {
	name string
	recv reflect.Value
	methods map[string]*methodType
}

func (m *methodType) newArgv() reflect.Value {
	var argv reflect.Value
	if m.ArgType.Kind() == reflect.Ptr {
		argv = reflect.New(m.ArgType.Elem()).Elem()
	} else {
		argv = reflect.New(m.ArgType).Elem()
	}
	return argv
}

func (m *methodType) newReplyv() reflect.Value {
	// ReplyType must be a pointer pointing to a struct
	replyv := reflect.New(m.ReplyType.Elem())
	return replyv
}

// create service and register service methods
func NewService(recv interface{}) *service {
	s := new(service)
	s.recv = reflect.ValueOf(recv)
	recvType := s.recv.Type()
	s.name = reflect.Indirect(reflect.ValueOf(recv)).Type().Name()
	s.methods = make(map[string]*methodType)

	for i:=0; i<recvType.NumMethod(); i++{
		method := recvType.Method(i)
		methodTyp := method.Type
		if methodTyp.NumIn() != 3 || methodTyp.NumOut() != 1 || !checkType(methodTyp) {
			continue
		}
		if methodTyp.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
			continue
		}
		argType, replyType := methodTyp.In(1), methodTyp.In(2)
		if !checkType(argType) || !checkType(replyType) {
			continue
		}
		s.methods[method.Name] = &methodType{
			method: method,
			ArgType: argType,
			ReplyType: replyType,
		}
	}
	return s
}

func checkType(t reflect.Type) bool{
	return ast.IsExported(t.Name()) || t.PkgPath() == ""
}

func (s *service) getMethod(serviceMethod string) *methodType {
	return s.methods[serviceMethod]
}

func (s *service) call(methodTyp *methodType, argv, replyv reflect.Value) error {
	f := methodTyp.method.Func
	returnValues := f.Call([]reflect.Value{s.recv, argv, replyv})
	if err := returnValues[0].Interface(); err != nil {
		return err.(error)
	}
	return nil
}