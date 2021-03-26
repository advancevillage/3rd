package ecies

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"net"
	"net/url"
	"strconv"
)

type IENode interface {
	GetId() []byte
	GetTcpPort() int
	GetUdpPort() int
	GetTcpHost() string
	GetPubKey() *ecdsa.PublicKey
	GetPriKey() *ecdsa.PrivateKey
	GetENodeUrl() (string, error)
}

//@overview: node, learn form eth.
//@author: richard.sun
//In the following example, the node URL describes
//a node with IP address 10.3.58.6, TCP listening port 30303
//and UDP discovery port 30301.
//
//   enode://<hex public key>@10.3.58.6:30303?discport=30301
//
type enode struct {
	tcpPort int
	udpPort int
	tcpHost net.IP
	pub     *ecdsa.PublicKey
	pri     *ecdsa.PrivateKey
}

func parse(rawUrl string) (*enode, error) {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}
	//1. 解析协议
	if u.Scheme != "enode" {
		return nil, fmt.Errorf("%s is invalid scheme", rawUrl)
	}
	//2. 解析公钥
	if u.User == nil {
		return nil, fmt.Errorf("%s is invalid public key", rawUrl)
	}
	b, err := hex.DecodeString(u.User.String())
	if err != nil {
		return nil, fmt.Errorf("%s is invalid hex public key", rawUrl)
	}
	x, y := elliptic.Unmarshal(elliptic.P256(), b)
	pub := &ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}
	//3. 解析IP
	ip := net.ParseIP(u.Hostname())
	if ip == nil {
		return nil, fmt.Errorf("%s is invalid ip", rawUrl)
	}
	//4. 解析TCP Port
	tcpPort, err := strconv.ParseUint(u.Port(), 10, 16)
	if err != nil {
		return nil, fmt.Errorf("%s is invalid tcp port", rawUrl)
	}
	//5. 解析UDP Port
	udpPort := tcpPort
	qv := u.Query()
	if len(qv.Get("discport")) > 0 {
		udpPort, err = strconv.ParseUint(qv.Get("discport"), 10, 16)
	}
	if err != nil {
		return nil, fmt.Errorf("%s is invalid udp port", rawUrl)
	}
	return &enode{tcpPort: int(tcpPort), tcpHost: ip, udpPort: int(udpPort), pub: pub}, nil
}

func (e *enode) raw() string {
	//1. hex public key
	var b = elliptic.Marshal(e.pub.Curve, e.pub.X, e.pub.Y)
	if e.tcpPort == e.udpPort {
		return fmt.Sprintf("enode://%x@%s:%d", b, e.tcpHost.String(), e.tcpPort)
	} else {
		return fmt.Sprintf("enode://%x@%s:%d?discport=%d", b, e.tcpHost.String(), e.tcpPort, e.udpPort)
	}
}

func (e *enode) id() []byte {
	var rawUrl = e.raw()
	var hash = sha256.New()
	hash.Write([]byte(rawUrl))
	return hash.Sum(nil)
}

func (e *enode) GetId() []byte {
	return e.id()
}

func (e *enode) GetTcpHost() string {
	return e.tcpHost.String()
}

func (e *enode) GetTcpPort() int {
	return e.tcpPort
}

func (e *enode) GetUdpPort() int {
	return e.udpPort
}

func (e *enode) GetPubKey() *ecdsa.PublicKey {
	return e.pub
}

func (e *enode) GetPriKey() *ecdsa.PrivateKey {
	return e.pri
}

func (e *enode) GetENodeUrl() (string, error) {
	return e.raw(), nil
}

func NewENodeByUrl(raw string) (IENode, error) {
	return parse(raw)
}

func NewENodeByPem(ppem []byte, host string, tcpPort int, udpPort int) (IENode, error) {
	//1. 解析私钥
	block, _ := pem.Decode(ppem)
	//2. x509
	pri, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("x509 parse private key %s", err.Error())
	}
	//3. 检查host
	ip := net.ParseIP(host)
	if ip == nil {
		return nil, fmt.Errorf("param %s host is invalid", host)
	}
	//4. 检查端口
	if tcpPort < 0 || tcpPort > 65535 || udpPort < 0 || udpPort > 65535 {
		return nil, fmt.Errorf("param %d or %d port is invalid", tcpPort, udpPort)
	}
	return &enode{tcpPort: tcpPort, udpPort: udpPort, tcpHost: ip, pub: &pri.PublicKey, pri: pri}, nil
}

func NewENodeByPP(pp *ecdsa.PrivateKey, host string, tcpPort int, udpPort int) (IENode, error) {
	//1. 检查host
	ip := net.ParseIP(host)
	if ip == nil {
		return nil, fmt.Errorf("param %s host is invalid", host)
	}
	if nil == pp {
		return nil, fmt.Errorf("ecdhe private key is invalid")
	}
	//2. 检查端口
	if tcpPort < 0 || tcpPort > 65535 || udpPort < 0 || udpPort > 65535 {
		return nil, fmt.Errorf("param %d or %d port is invalid", tcpPort, udpPort)
	}
	return &enode{tcpPort: tcpPort, udpPort: udpPort, tcpHost: ip, pub: &pp.PublicKey, pri: pp}, nil
}

func NewENodeByPub(pub *ecdsa.PublicKey) (IENode, error) {
	return &enode{
		tcpPort: 0,
		udpPort: 0,
		tcpHost: nil,
		pub:     pub,
		pri:     nil,
	}, nil
}
