package utils

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"math"
	"os"
)

// Md5SmallFile 文件求md5值 - 小文件
func Md5SmallFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	h := md5.New()
	_, err = io.Copy(h, file)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// Md5BigFile 文件求md5值 - 大文件
func Md5BigFile(path string) (string, error) {
	var fileChunk uint64 = 10485760
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	// calculate the file size
	info, _ := file.Stat()
	fileSize := info.Size()
	blocks := uint64(math.Ceil(float64(fileSize) / float64(fileChunk)))
	h := md5.New()
	for i := uint64(0); i < blocks; i++ {
		blockSize := int(math.Min(float64(fileChunk), float64(fileSize-int64(i*fileChunk))))
		buf := make([]byte, blockSize)
		_, err = file.Read(buf)
		if err != nil {
			return "", err
		}
		_, err = io.WriteString(h, string(buf)) // append into the hash
		if err != nil {
			return "", err
		}
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// MD5String 字符串md5值
func MD5String(v string) string {
	d := []byte(v)
	m := md5.New()
	m.Write(d)
	return hex.EncodeToString(m.Sum(nil))
}
