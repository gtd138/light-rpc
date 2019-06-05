package lrpc

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"

	"github.com/astaxie/beego/logs"
)

type RPCCall struct {
	Seq    int
	Class  string
	Method string
	Args   []interface{}
}

type RPCResult struct {
	Seq    int
	Return []interface{}
}

// 处理字节
func handleResultByte(bytes []byte) (buf []byte, results []*RPCResult) {
	for {
		length := len(bytes)
		if length <= 0 {
			//beego.Error("HandleReceive:消息长度小于零，停止解析。。。")
			break
		}

		// 头的长度必须为4个字节
		if len(bytes) < 4 {
			buf = bytes
			break
		}

		// 处理一下头
		bytes := HandleHeader(bytes)
		if len(bytes) < 8 {
			buf = bytes
			break
		}

		byteLen := GetMsgLen(bytes)
		if byteLen == 0 {
			buf = bytes
			break
		}

		body := GetMsgBody(bytes, byteLen)
		if body == nil {
			buf = bytes
			break
		}

		// 先缓把接收的字节缓存一下
		pack := DecodeResultMsg(body)
		if pack.Seq != 0 {
			if results == nil {
				results = make([]*RPCResult, 0)
			}
			results = append(results, &pack)
		} else {
			logs.Error("HandleReceive:反序列化body失败!")
			buf = bytes
			break
		}

		// 裁剪buffer
		msgLen := 8 + int(byteLen)
		lassLen := len(bytes) - msgLen
		if lassLen <= 0 {
			break
		} else {
			bytes = bytes[msgLen:]
		}
	}

	return
}

// 处理字节
func handleCallByte(bytes []byte) (buf []byte, results []*RPCCall) {
	for {
		length := len(bytes)
		if length <= 0 {
			//beego.Error("HandleReceive:消息长度小于零，停止解析。。。")
			break
		}

		// 头的长度必须为4个字节
		if len(bytes) < 4 {
			buf = bytes
			break
		}

		// 处理一下头
		bytes := HandleHeader(bytes)
		if len(bytes) < 8 {
			buf = bytes
			break
		}

		byteLen := GetMsgLen(bytes)
		if byteLen == 0 {
			buf = bytes
			break
		}

		body := GetMsgBody(bytes, byteLen)
		if body == nil {
			buf = bytes
			break
		}

		// 先缓把接收的字节缓存一下
		pack := DecodeCallMsg(body)
		if pack.Seq != 0 {
			if results == nil {
				results = make([]*RPCCall, 0)
			}
			results = append(results, &pack)
		} else {
			logs.Error("HandleReceive:反序列化body失败!")
			buf = bytes
			break
		}

		// 裁剪buffer
		msgLen := 8 + int(byteLen)
		lassLen := len(bytes) - msgLen
		if lassLen <= 0 {
			break
		} else {
			bytes = bytes[msgLen:]
		}
	}

	return
}

// 处理一下消息头
func HandleHeader(b []byte) (newB []byte) {
	head := b[0:4]
	//如果消息头不正确，丢弃此段消息，并找出正确的
	if !bytes.Equal(head, MSG_HEAD) {
		//b = b[4:]
		for len(b) > 0 {
			if b[0] != MSG_HEAD[0] {
				b = b[1:]
				continue
			}

			if len(b) < 2 {
				break
			}

			if b[1] != MSG_HEAD[1] {
				b = b[2:]
				continue
			}

			if len(b) < 3 {
				break
			}

			if b[2] != MSG_HEAD[2] {
				b = b[3:]
				continue
			}

			if len(b) < 4 {
				break
			}

			if b[3] != MSG_HEAD[3] {
				b = b[4:]
				continue
			}

			break
		}
	}

	newB = b

	return
}

// 获取长度
func GetMsgLen(b []byte) uint {
	if len(b) <= 8 {
		return 0
	}

	lenByte := b[4:8]
	return BytesToIntLittleEndian(lenByte)
}

// 字节转换成整形(小端)
func BytesToIntLittleEndian(b []byte) uint {
	bytesBuffer := bytes.NewBuffer(b)
	var tmp int32
	binary.Read(bytesBuffer, binary.LittleEndian, &tmp)
	return uint(tmp)
}

//整形转换成字节(小端)
func IntToBytesLittleEndian(n int) []byte {
	tmp := int32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.LittleEndian, tmp)
	return bytesBuffer.Bytes()
}

// 获取body
func GetMsgBody(b []byte, length uint) []byte {
	lessByte := b[8:]
	//beego.Debug("GetMsgBody:length = ", length)
	if uint(len(lessByte)) >= length {
		body := b[8 : length+8]
		return body
	}
	return nil
}

// 反序列化消息
func DecodeResultMsg(b []byte) (rpccall RPCResult) {
	buffer := bytes.NewBuffer(b)
	dec := gob.NewDecoder(buffer)
	rpccall = RPCResult{}
	dec.Decode(&rpccall)
	return
}

// 反序列化消息
func DecodeCallMsg(b []byte) (rpccall RPCCall) {
	buffer := bytes.NewBuffer(b)
	dec := gob.NewDecoder(buffer)
	rpccall = RPCCall{}
	dec.Decode(&rpccall)
	return
}

// 序列化消息
func EncodeResultMsg(in *RPCResult) (b []byte) {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(in)
	if err != nil {
		fmt.Println("error: ", err)
		return nil
	}

	b = buffer.Bytes()
	return
}

// 序列化消息
func EncodeCallMsg(in *RPCCall) (b []byte) {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(in)
	if err != nil {
		fmt.Println("error: ", err)
		return nil
	}

	b = buffer.Bytes()
	return
}

// 组合成消息
func AssembleBuffer(b []byte) (result []byte) {
	if b == nil || len(b) <= 0 {
		return
	}

	var lenByte []byte
	var bLen = len(b)
	lenByte = IntToBytesLittleEndian(bLen)

	buffer := bytes.NewBuffer(MSG_HEAD)
	buffer.Write(lenByte)

	buffer.Write(b)
	result = buffer.Bytes()

	return
}
