package netx

import "io"

type IProtocol interface {
	HeaderLength() int32
	Write(io.Writer, []byte)
}

//@overview: header + body 自定义协议. 粘包和拆包问题应用广泛的解决思路.
//HBProtocol is Header and Body Protocol
//@author: richard.sun
type HBProtocol struct {
	hLen int32
	wf   func(io.Writer, []byte)
}

func NewHBProtocol(hLen int32, wf func(io.Writer, []byte)) IProtocol {
	return &HBProtocol{hLen: hLen, wf: wf}
}

func (p *HBProtocol) HeaderLength() int32 {
	return p.hLen
}

func (p *HBProtocol) Write(w io.Writer, b []byte) {
	p.wf(w, b)
}
