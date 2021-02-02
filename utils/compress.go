package utils

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
