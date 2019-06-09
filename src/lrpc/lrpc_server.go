package lrpc

import (
	"net"
	"reflect"
	"sync"

	"log"
)

var MSG_HEAD = []byte("@@@@")

// rpc
type RPCServer struct {
	Listener *net.TCPListener
	Port     string
	RPCObjs  *sync.Map // rpc注册的对象
}

func NewRPCServer(port string) *RPCServer {
	return &RPCServer{
		Port:    port,
		RPCObjs: &sync.Map{},
	}
}

// 注册类
func (this *RPCServer) Register(class interface{}) {
	value := reflect.ValueOf(class)
	name := reflect.Indirect(value).Type().Name()

	this.RPCObjs.Store(name, class)
}

// 接受消息
func (this *RPCServer) Accept() {
	tcpAddr, err := net.ResolveTCPAddr("tcp", ":"+this.Port)
	if err != nil {
		str := "error:" + err.Error()
		panic(str)
	}

	this.Listener, err = net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		str := "error:" + err.Error()
		panic(str)
	}

	for {
		conn, err := this.Listener.Accept()
		if err != nil {
			//logs.Error("RPCError:",)
			log.Println("error: ", err)
			if conn != nil {
				conn.Close()
			}
			continue
		}

		go this.HandleConn(conn)
	}
}

// 处理链接
func (this *RPCServer) HandleConn(conn net.Conn) {
	buffer := make([]byte, 0)
	for {
		req := make([]byte, 512)
		n, err := conn.Read(req)
		if err == nil && n > 0 {
			b := req[:n]
			b = append(buffer, b...)
			var calls []*RPCCall
			buffer, calls = handleCallByte(b)
			if calls != nil && len(calls) > 0 {
				results := this.handleFunc(calls)
				this.handleResult(results, conn)
			}
		} else {
			log.Println("error = ", err)
			break
		}
	}

	conn.Close()
}

// 调用函数
func (this *RPCServer) handleFunc(calls []*RPCCall) (results []*RPCResult) {
	for i := range calls {
		call := calls[i]

		if e, ok := this.RPCObjs.Load(call.Class); ok {
			obj := reflect.ValueOf(e)
			var params []reflect.Value
			if call.Args != nil {
				params = make([]reflect.Value, 0)
				for i := range call.Args {
					v := reflect.ValueOf(call.Args[i])
					params = append(params, v)
				}
			}

			method := obj.MethodByName(call.Method)
			values := method.Call(params)
			var r *RPCResult
			if values != nil && len(values) > 0 {
				r = &RPCResult{
					Seq:    call.Seq,
					Return: make([]interface{}, 0),
				}

				for j := range values {
					r.Return = append(r.Return, values[j].Interface())
				}

				if results == nil {
					results = make([]*RPCResult, 0)
				}

				results = append(results, r)
			}
		} else {
			log.Println("obj is not exiset!")
		}
	}

	return
}

// 处理rpc结果
func (this *RPCServer) handleResult(results []*RPCResult, conn net.Conn) {
	// 没有返回值不处理
	if results == nil || len(results) <= 0 {
		return
	}

	for i := range results {
		r := results[i]
		b := EncodeResultMsg(r)
		if b == nil {
			continue
		}
		b = AssembleBuffer(b)
		if b == nil {
			continue
		}
		_, err := conn.Write(b)
		if err != nil {
			log.Println("error:", err)
		}
	}
}
