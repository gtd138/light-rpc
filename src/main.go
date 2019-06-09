package main

import (
	"fmt"
	"time"

	"./lrpc"
)

var count = 0
var startTime int64 = 0
var client *lrpc.RPCClient

// 生成时间戳
func GetCurrentTimeStampMS() int64 {
	return time.Now().UnixNano() / 1e6
}

type Test struct {
}

func (this *Test) Add(a, b int) int {
	return a + b
}

func (this *Test) Hello() {
	fmt.Println("hello!")
}

func (this *Test) GetName() (string, int) {
	return "Peter", 20
}

func main() {
	go Server()

	startTime = GetCurrentTimeStampMS()
	go Client()

	for {
		a := 1
		a++
		time.Sleep(time.Second)
	}
}

func Client() {
	if client == nil {
		client = lrpc.NewRPCClient("localhost", "2048")
		client.Dial()
	}

	client.CallReply("Test.Add", []interface{}{1, 2}, func(arg ...interface{}) {
		count++
		if count < 10000 {
			Client()
		} else {
			dtime := GetCurrentTimeStampMS() - startTime
			fmt.Println("dTime:", dtime)

			client.Call("Test.Hello", nil)

			client.CallReply("Test.GetName", nil, func(arg ...interface{}) {
				fmt.Println(arg)
			})
		}
	})
}

func Server() {
	server := lrpc.NewRPCServer("2048")
	server.Register(&Test{})
	server.Accept()
}
