package funker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"reflect"
)

// Handle a Funker function.
func Handle(handler interface{}) error {
	handlerValue := reflect.ValueOf(handler)
	handlerType := handlerValue.Type()
	if handlerType.Kind() != reflect.Func || handlerType.NumIn() != 1 || handlerType.NumOut() != 1 {
		return fmt.Errorf("Handler must be a function with a single parameter and single return value.")
	}
	argsValue := reflect.New(handlerType.In(0))

	// TODO (bfirsh): no easy way to set connection queue to 1, yet.
	// https://github.com/golang/go/issues/9661
	listener, err := net.Listen("tcp", ":9999")
	if err != nil {
		return err
	}
	conn, err := listener.Accept()
	if err != nil {
		return err
	}
	argsJSON, err := ioutil.ReadAll(conn)
	if err != nil {
		return err
	}
	err = json.Unmarshal(argsJSON, argsValue.Interface())
	if err != nil {
		return err
	}

	ret := handlerValue.Call([]reflect.Value{argsValue.Elem()})[0].Interface()
	retJSON, err := json.Marshal(ret)
	if err != nil {
		return err
	}

	if _, err = conn.Write(retJSON); err != nil {
		return err
	}

	if err = conn.Close(); err != nil {
		return err
	}

	return listener.Close()
}
