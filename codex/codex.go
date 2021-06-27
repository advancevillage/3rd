package codex

import (
	"encoding/json"
	"errors"
)

type CodeType uint8

const (
	Bencode = CodeType(0x00)
	Rlp     = CodeType(0x10)
	Json    = CodeType(0x20)
	Gob     = CodeType(0x30)
)

var codeTypeErr = errors.New("don't support code type")

func Marshal(t CodeType, v interface{}) ([]byte, error) {
	switch t {
	case Json:
		return json.Marshal(v)
	default:
		return nil, codeTypeErr
	}
}

func Unmarshal(t CodeType, data []byte, v interface{}) error {
	switch t {
	case Json:
		return json.Unmarshal(data, v)
	default:
		return codeTypeErr
	}
}
