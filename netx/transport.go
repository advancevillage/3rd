package netx

import "golang.org/x/net/context"

//@overview: 传输层接口协议. 特点是安全,灵活,高效
//@author: richard.sun
//@note:
//1. write frame
//2. read frame
type ITransport interface {
	ReadFrame(context.Context) ([]byte, error)
	WriteFrame(context.Context, []byte) error
}

type TransportOption struct {
	MaxSize uint32
}

type tcpConn struct {
}

//message auth code
type tcpMac struct {
}
