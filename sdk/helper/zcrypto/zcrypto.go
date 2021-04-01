package zcrypto

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"io"
)

//文件头
type FileHeader struct {
	Name []byte //加密后的原文件名
	IV   []byte //初始化向量
	Rand [64]byte
	Ctr  []byte
}

func readBlock(src io.Reader) (io.Reader, error) {
	hsize := make([]byte, 4)
	n, err := io.ReadFull(src, hsize)
	if err != nil {
		return nil, err
	}
	if n != 4 {
		return nil, errors.New("wrong format")
	}
	sz1 := binary.BigEndian.Uint32(hsize)
	buf := make([]byte, sz1)
	n, err = io.ReadFull(src, buf)
	if err != nil {
		return nil, err
	}
	if uint32(n) != sz1 {
		return nil, errors.New("wrong format")
	}
	return bytes.NewBuffer(buf), nil
}

func writeBlock(data []byte, dst io.Writer) error {
	sz1 := len(data)
	hsize := make([]byte, 4)
	binary.BigEndian.PutUint32(hsize, uint32(sz1))
	_, err := dst.Write(hsize)
	if err != nil {
		return err
	}
	_, err = dst.Write(data)
	if err != nil {
		return err
	}
	return nil
}

//生成32位密码
func getShaKey(pwd string) []byte {
	if len(pwd) == 0 {
		return nil
	}
	bp := []byte(pwd)
	buf := bytes.NewBuffer(make([]byte, 0, len(bp)*40))
	for i := 0; i < 40; i++ {
		buf.Write(bp)
	}

	res := sha256.Sum256(buf.Bytes())

	return res[:]
}