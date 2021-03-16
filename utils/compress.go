package utils

import (
	"bytes"
)

type ICompress interface {
	Compress([]byte) ([]byte, error)
	Uncompress([]byte) ([]byte, error)
}

//@overview: rle
//@auther: richard.sun
//@note:
//1.  添加一个标志字节块
//	  高位 = 1  低7位表示 下一个字节的重复个数
//	  高位 = 0  低7位表示 下一个不重复数据块的长度
//2. 当初重复数据块长度或者不重复数据块长度大于127 需要分组
//3. 重复字节块表示重复字节长度大于2时
type rle struct{}

func NewRLE() (ICompress, error) {
	return &rle{}, nil
}

//http://paulbourke.net/dataformats/compress/rle.c
func (r *rle) Compress(s []byte) ([]byte, error) {
	var (
		ss     []byte
		item   byte
		count  int
		index  int
		length = len(s)
	)
	for count < length {
		index = count
		item = s[index]
		index++
		for index < length && index-count < 0x7f && item == s[index] {
			index++
		}
		//至少元素重复3次
		if index-count < 0x3 {
			for index < length && index-count < 0x7f && (s[index] != s[index-1] || index > 1 && s[index] != s[index-2]) {
				index++
			}
			for index < length && s[index] == s[index-1] {
				index--
			}
			ss = append(ss, 0x7f&byte(index-count))
			ss = append(ss, s[count:index]...)
		} else {
			ss = append(ss, 0x80|byte(index-count))
			ss = append(ss, item)
		}
		count = index
	}
	return ss, nil
}

func (r *rle) Uncompress(s []byte) ([]byte, error) {
	var (
		length = len(s)
		ss     []byte
	)
	for length > 0 {
		switch {
		case s[0] == 0x80:
			s = s[1:]
			length = len(s)
		case s[0] > 0x80:
			ss = append(ss, bytes.Repeat([]byte{s[1]}, int(s[0]&0x7f))...)
			length -= 0x2
			s = s[0x2:]
		case s[0] < 0x80:
			ss = append(ss, s[0x1:0x1+s[0]&0x7f]...)
			length -= 0x01 + int(s[0]&0x7f)
			s = s[0x1+s[0]&0x7f:]
		}
	}
	return ss, nil
}
