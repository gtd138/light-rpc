package lrpc

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"
)

type RPCFunc func(arg ...interface{})

type RPCClient struct {
	count   int32
	addr    string
	port    string
	buffer  []byte
	conn    net.Conn
	funcMap *sync.Map
}

func NewRPCClient(addr, port string) *RPCClient {
	return &RPCClient{
		count:   0,
		buffer:  make([]byte, 0),
		addr:    addr,
		port:    port,
		funcMap: &sync.Map{},
	}
}

func (this *RPCClient) Dial() {
	var err error
	this.conn, err = net.Dial("tcp", this.addr+":"+this.port)
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			b := make([]byte, 512)
			n, err := this.conn.Read(b)
			if err == nil {
				if n <= 0 {
					continue
				}
				b = b[:n]
				var results []*RPCResult
				this.buffer, results = handleResultByte(b)
				if results != nil {
					this.handleResults(results)
				}
			} else {
				fmt.Println(err)
				break
			}
		}

		this.conn.Close()
	}()
}

// 处理消息
func (this *RPCClient) handleResults(results []*RPCResult) {
	for i := range results {
		r := results[i]
		if e, ok := this.funcMap.Load(r.Seq); ok {
			fn := e.(RPCFunc)
			if r.Return != nil {
				fn(r.Return...)
			} else {
				fn()
			}
			this.funcMap.Delete(r.Seq)
		}
	}
}

// 远程调用(无返回值)
// method: class.method
func (this *RPCClient) Call(method string, args []interface{}) {
	this.CallReply(method, args, nil)
}

// 远程调用(有返回值)
// method: class.method
func (this *RPCClient) CallReply(method string, args []interface{}, fn RPCFunc) {
	m := strings.Split(method, ".")
	atomic.AddInt32(&this.count, 1)
	call := &RPCCall{
		Class:  m[0],
		Method: m[1],
		Args:   args,
		Seq:    int(atomic.LoadInt32(&this.count)),
	}

	if fn != nil {
		this.funcMap.Store(call.Seq, fn)
	}

	b := EncodeCallMsg(call)
	b = AssembleBuffer(b)

	this.conn.Write(b)
}
