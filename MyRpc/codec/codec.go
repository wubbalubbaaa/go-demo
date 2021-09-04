package codec

import (
	"io"
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
// NewCodecFunc 抽象出 Codec 的构造函数，客户端和服务端可以通过 Codec 的 Type 得到构造函数，从而创建 Codec 实例。
// 这部分代码和工厂模式类似，与工厂模式不同的是，返回的是构造函数，而非实例
type NewCodecFunc func(io.ReadWriteCloser) Codec

type Type string

const (
	GobType  Type = "application/gob"
	JsonType Type = "application/json" // not implemented
)

var NewCodecFuncMap map[Type]NewCodecFunc

func init() {
	NewCodecFuncMap = make(map[Type]NewCodecFunc)
	NewCodecFuncMap[GobType] = NewGobCodec
}
