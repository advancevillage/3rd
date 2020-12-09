package netx

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type ProtocolRequest struct {
	Body []byte
}

type ProtocolResponse struct {
	Body []byte
}

type ProtocolFunc func(*ProtocolRequest) (*ProtocolResponse, error)

type IProtocol interface {
	ReadHeader(*bufio.Reader) (int32, error)
	ReadBody(*bufio.Reader, int32) ([]byte, error)
	Write(*bufio.Writer, []byte) error
}

//@overview: header + body 自定义协议. 粘包和拆包问题应用广泛的解决思路.
//HBProtocol is Header and Body Protocol
//@author: richard.sun
type HBProtocol struct {
	hLen int32
	f    ProtocolFunc
}

func NewHBProtocol(hLen int32, f ProtocolFunc) IProtocol {
	return &HBProtocol{hLen: hLen, f: f}
}

func (p *HBProtocol) Write(w *bufio.Writer, b []byte) error {
	var req = &ProtocolRequest{Body: b}
	//1. 协议处理函数
	var res, err = p.f(req)
	if err != nil {
		return err
	}
	//2. 构造响应头信息
	var h = make([]byte, p.hLen)
	binary.BigEndian.PutUint32(h, uint32(len(res.Body)))
	//3. 构造响应体
	var body = append(h, res.Body...)
	fmt.Println(body)
	//4. 响应
	_, err = w.Write(body)
	if err != nil {
		return err
	}
	return w.Flush()
}

func (p *HBProtocol) ReadHeader(r *bufio.Reader) (int32, error) {
	//1. 解析协议Header
	var hLen = p.hLen
	var bLen = int32(0)
	var h, err = r.Peek(int(hLen))
	if err == io.EOF {
		return 0, io.EOF
	}
	if err != nil {
		return 0, err
	}
	//2. 解析Body长度
	err = binary.Read(bytes.NewBuffer(h), binary.BigEndian, &bLen)
	if err != nil {
		return 0, err
	}
	//3. 是否是完成的包
	if int32(r.Buffered()) < (hLen + bLen) {
		return 0, ErrPartPackage
	}
	return hLen + bLen, nil
}

func (p *HBProtocol) ReadBody(r *bufio.Reader, n int32) ([]byte, error) {
	var b = make([]byte, n)
	var nn, err = r.Read(b)
	if err != nil {
		return nil, err
	}
	return b[p.hLen:nn], nil
}
