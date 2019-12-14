//author: richard
package utils

import (
	"crypto/aes"
	"crypto/cipher"
)

//aes最少以16个字节分组，密钥key和随机数种子Iv和分组字节数相同是16个字节
const (
 	sign  = "kelly68chen.1995"
	iv = "1995.chen86kelly"
)

//@link: https://juejin.im/post/5d2b0fbf51882547b2361a8a
func EncryptUseAes(plain []byte) ([]byte, error) {
	var block cipher.Block
	var err error
	if block, err = aes.NewCipher([]byte(sign)); err != nil{
		return nil, err
	}
	//创建ctr
	stream := cipher.NewCTR(block, []byte(iv))
	//加密, src,dst 可以为同一个内存地址
	stream.XORKeyStream(plain, plain)
	return plain, nil
}

func DecryptUseAes(cipher []byte) ([]byte, error){
	//对密文再进行一次按位异或就可以得到明文
	//例如：3的二进制是0011和8的二进制1000按位异或(相同为0,不同为1)后得到1011，
	//对1011和8的二进制1000再进行按位异或得到0011即是3
	return EncryptUseAes(cipher)
}