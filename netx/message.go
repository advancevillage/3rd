package netx

import (
	"errors"

	"github.com/advancevillage/3rd/codex"
	"github.com/advancevillage/3rd/utils"
)

type IMessage interface {
	Decode(b []byte, t *MessageType, seq *uint16, v interface{}) error
	Encode(t MessageType, seq uint16, v interface{}) ([]byte, error)
}

const headerSize = 8

var (
	messageErr         = errors.New("message error")
	messageLenErr      = errors.New("message length out of bounds")
	messageHeaderErr   = errors.New("message header size small")
	messageChecksumErr = errors.New("message checksum error")
)

type MessageType uint8

type message struct {
	msgFlag byte //消息控制 高4位是编码方式,低4位保留
	data    []byte
}

func newMsgCli(c codex.CodeType) IMessage {
	return &message{
		msgFlag: 0xf0 & uint8(c),
	}
}

func (m *message) Decode(b []byte, t *MessageType, seq *uint16, v interface{}) error {
	if len(b) < headerSize {
		return messageHeaderErr
	}
	var mm = new(message)
	mm.data = b

	var l = m.readLength()
	var s = m.readChecksum()
	if int(l) > len(b) {
		return messageErr
	}
	var s2 = utils.CRC16(b[headerSize:l])
	if s != s2 {
		return messageChecksumErr
	}
	*t = MessageType(m.readMsgType())
	*seq = m.readMsgId()
	err := codex.Unmarshal(codex.CodeType(m.msgFlag), b[headerSize:l], v)
	if err != nil {
		return err
	}
	return nil
}

func (m *message) Encode(t MessageType, seq uint16, v interface{}) ([]byte, error) {
	//1. 数据编码
	var payload, err = codex.Marshal(codex.CodeType(m.msgFlag), v)
	if err != nil {
		return nil, err
	}
	if len(payload)+headerSize > (1 << 16) {
		return nil, messageLenErr
	}
	//2. 构造message
	var mm = new(message)
	mm.data = make([]byte, headerSize+len(payload))

	mm.writePayload(payload)
	mm.writeType(uint8(t))
	mm.writeFlag(uint8(m.msgFlag))
	mm.writeLength(uint16(headerSize + len(payload)))
	mm.writeMsgId(seq)

	return mm.data, nil
}

func (m *message) writeLength(v uint16) {
	m.data[2] = byte(v >> 8)
	m.data[3] = byte(v)
}

func (m *message) readLength() uint16 {
	return uint16(m.data[2]) | uint16(m.data[3])<<8
}

func (m *message) writeType(v uint8) {
	m.data[0] = 0xff & v
}

func (m *message) readMsgType() uint8 {
	return uint8(m.data[0])
}

func (m *message) writeFlag(v uint8) {
	m.data[1] = 0xff & v
}

func (m *message) writeMsgId(v uint16) {
	m.data[4] = byte(v >> 8)
	m.data[5] = byte(v)
}

func (m *message) readMsgId() uint16 {
	return uint16(m.data[4]) | uint16(m.data[5])<<8
}

func (m *message) writeChecksum() {
	//1. compute crc16
	var crc16 = utils.CRC16(m.data[headerSize:])
	//2. write checksum
	m.data[6] = byte(crc16 >> 8)
	m.data[7] = byte(crc16)
}

func (m *message) readChecksum() uint16 {
	return uint16(m.data[6]) | uint16(m.data[7])<<8
}

func (m *message) writePayload(v []byte) {
	copy(m.data[headerSize:], v)
}
