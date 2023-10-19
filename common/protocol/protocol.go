package protocol

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
)

// Sign 数据流的签名
// TODO 用于安全, 签名改为由master进行颁发
var Sign = "268ce0d3010f8d22"

// Stream UDP数据包结构
type Stream struct {
	Sign    string      // 签名16位
	CtxId   string      // 上下文id  8位
	Command CommandCode // 命令代码  1位
	Data    []byte      // 数据
}

// Packet 封包
func Packet(cmd CommandCode, ctxId string, data []byte) ([]byte, error) {
	var (
		err    error
		stream []byte
		buf    = new(bytes.Buffer)
	)
	if len(Sign) != 16 {
		return stream, fmt.Errorf("非法的签名")
	}
	_ = binary.Write(buf, binary.LittleEndian, []byte(Sign))
	if len(ctxId) != 16 {
		_ = binary.Write(buf, binary.LittleEndian, []byte("0000000000000000"))
	} else {
		_ = binary.Write(buf, binary.LittleEndian, []byte(ctxId))
	}
	_ = binary.Write(buf, binary.LittleEndian, cmd)
	err = binary.Write(buf, binary.LittleEndian, data)
	if err != nil {
		return stream, err
	}
	//log.Println(buf.Bytes())
	if buf.Len() < 33 {
		return stream, fmt.Errorf("非法数据!")
	}
	stream = buf.Bytes()
	return stream, nil
}

// DecryptPacket 解包
func DecryptPacket(data []byte, n int) *Stream {
	a := data[32:33]
	stream := &Stream{
		Sign:    string(data[0:16]),
		CtxId:   string(data[16:32]),
		Command: CommandCode(uint8(a[0])),
		Data:    data[33:n],
	}
	return stream
}

// GobEncoder 使用 god进行包压缩
func GobEncoder(data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(data)
	return buf.Bytes(), err
}

// GobDecoder 使用 god进行包解压
func GobDecoder(obj interface{}, stream []byte) error {
	var data bytes.Buffer
	data.Write(stream)
	dec := gob.NewDecoder(&data)
	return dec.Decode(obj)
}

// GzipCompress gzip压缩
func GzipCompress(src []byte) []byte {
	var in bytes.Buffer
	w, _ := gzip.NewWriterLevel(&in, gzip.BestCompression)
	//w := gzip.NewWriter(&in)
	_, _ = w.Write(src)
	_ = w.Close()
	return in.Bytes()
}

// GzipDecompress gzip解压
func GzipDecompress(src []byte) ([]byte, error) {
	reader := bytes.NewReader(src)
	gr, err := gzip.NewReader(reader)
	if err != nil {
		return []byte(""), err
	}
	bf := make([]byte, 0)
	buf := bytes.NewBuffer(bf)
	_, err = io.Copy(buf, gr)
	err = gr.Close()
	return buf.Bytes(), nil
}

// Struct2Byte 结构体转字节
func Struct2Byte(obj interface{}) ([]byte, error) {
	return json.Marshal(obj)
}

// DataEncoder 数据量大，使用 json 序列化+gzip压缩
func DataEncoder(obj interface{}) ([]byte, error) {
	b, err := json.Marshal(obj)
	if err != nil {
		return []byte(""), err
	}
	return GzipCompress(b), nil
}

// DataDecoder 解码
func DataDecoder(data []byte, obj interface{}) error {
	b, err := GzipDecompress(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, obj)
}
