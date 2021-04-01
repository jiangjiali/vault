package zcrypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/gob"
	"errors"
	"io"
	"os"
	"path/filepath"
)

func CfbEncryptoToFile(src, dst, pwd string) error {
	fp1, err := os.Open(src)
	if err != nil {
		return err
	}
	defer fp1.Close()

	fp2 := os.Stdout
	if dst != "-" {
		fp2, err = os.Create(dst)
		if err != nil {
			return err
		}
		defer fp2.Close()
	}

	h1, err := newCfbFileHeader(filepath.Base(src), pwd)
	if err != nil {
		return err
	}

	buf := bytes.NewBufferString("")
	gobEnc := gob.NewEncoder(buf)
	err = gobEnc.Encode(h1)
	if err != nil {
		return err
	}

	err = writeBlock(buf.Bytes(), fp2)
	if err != nil {
		return err
	}

	key := getShaKey(pwd)
	if key == nil {
		return errors.New("缺少密码")
	}

	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	aesEncoder := cipher.NewCFBEncrypter(aesBlock, h1.IV)
	wr := cipher.StreamWriter{S: aesEncoder, W: fp2}
	defer wr.Close()

	_, err = io.Copy(wr, fp1)
	if err != nil {
		return err
	}

	return nil
}

func CfbDecryptoFromFile(src, dst, pwd string) error {
	fp1, err := os.Open(src)
	if err != nil {
		return err
	}
	defer fp1.Close()

	fp2 := os.Stdout
	if dst != "-" {
		fp2, err = os.Create(dst)
		if err != nil {
			return err
		}
		defer fp2.Close()
	}

	rd1, err := readBlock(fp1)
	if err != nil {
		//log.Println(err)
		return err
	}

	_, iv, err := CfbReadName(rd1, pwd)
	if err != nil {
		//log.Println(err)
		return err
	}
	//log.Println(name)

	key := getShaKey(pwd)
	if key == nil {
		return errors.New("缺少密码")
	}

	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	aesStream := cipher.NewCFBDecrypter(aesBlock, iv)
	rd := cipher.StreamReader{S: aesStream, R: fp1}

	_, err = io.Copy(fp2, rd)

	return err
}

func CfbReadName(src io.Reader, pwd string) (string, []byte, error) {
	h1 := new(FileHeader)
	gobDec := gob.NewDecoder(src)
	err := gobDec.Decode(h1)
	if err != nil {
		//log.Println(err)
		return "", nil, err
	}
	//log.Printf("%x\n%x\n", h1.Name, h1.IV)

	key := getShaKey(pwd)
	if key == nil {
		return "", nil, errors.New("缺少密码")
	}

	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		return "", nil, err
	}

	iv := make([]byte, aesBlock.BlockSize())
	//decrypt Ctr , then compare with Rand
	copy(iv, h1.IV)
	aesStream := cipher.NewCFBDecrypter(aesBlock, iv)
	rd := cipher.StreamReader{S: aesStream, R: bytes.NewBuffer(h1.Ctr)}
	buf := bytes.NewBufferString("")
	_, err = io.Copy(buf, rd)
	if err != nil {
		return "", nil, err
	}
	if bytes.Compare(h1.Rand[:], buf.Bytes()) != 0 {
		return "", nil, errors.New("密码无效！")
	}

	//decrypt Name
	copy(iv, h1.IV)

	aesStream = cipher.NewCFBDecrypter(aesBlock, iv)
	rd = cipher.StreamReader{S: aesStream, R: bytes.NewBuffer(h1.Name)}
	res := bytes.NewBufferString("")
	_, err = io.Copy(res, rd)
	return res.String(), h1.IV, err
}

//从文件名读取原文件名
func CfbReadNameFromFile(src, pwd string) (string, error) {
	fp1, err := os.Open(src)
	if err != nil {
		return "", err
	}
	defer fp1.Close()

	rd1, err := readBlock(fp1)
	if err != nil {
		//log.Println(err)
		return "", err
	}

	name, _, err := CfbReadName(rd1, pwd)
	if err != nil {
		//log.Println(err)
		return "", err
	}
	return name, err
}

//初始化FileHeader
func newCfbFileHeader(name, pwd string) (*FileHeader, error) {
	if len(name) == 0 {
		return nil, errors.New("缺少文件名")
	}

	key := getShaKey(pwd)
	if key == nil {
		return nil, errors.New("缺少密码")
	}

	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	res := new(FileHeader)
	res.IV = make([]byte, aesBlock.BlockSize())
	iv := make([]byte, aesBlock.BlockSize())
	_, err = io.ReadFull(rand.Reader, iv)
	if err != nil {
		return nil, err
	}
	copy(res.IV, iv)

	buf := bytes.NewBufferString("")
	aesEncoder := cipher.NewCFBEncrypter(aesBlock, iv)
	wr := cipher.StreamWriter{S: aesEncoder, W: buf}
	_, err = wr.Write([]byte(name))
	if err != nil {
		return nil, err
	}
	wr.Close()

	res.Name = buf.Bytes()

	//set Rand and Crt
	_, err = io.ReadFull(rand.Reader, res.Rand[:])
	if err != nil {
		return nil, err
	}
	buf = bytes.NewBufferString("")
	copy(iv, res.IV)
	aesEncoder = cipher.NewCFBEncrypter(aesBlock, iv)
	wr = cipher.StreamWriter{S: aesEncoder, W: buf}
	_, err = wr.Write(res.Rand[:])
	if err != nil {
		return nil, err
	}
	wr.Close()
	res.Ctr = buf.Bytes()

	return res, nil
}
