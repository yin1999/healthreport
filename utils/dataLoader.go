package utils

import (
	"bytes"
	"encoding/gob"
	"os"
)

// DataStore Persist data to file
func DataStore(data interface{}, filename string) error { // 二进制存储，并做简单混淆
	buffer := new(bytes.Buffer)
	encoder := gob.NewEncoder(buffer)
	err := encoder.Encode(data)
	if err != nil {
		return err
	}
	raw := buffer.Bytes()
	byteGarble(raw)
	return os.WriteFile(filename, raw, 0400) // Read only(owner)
}

// DataLoad read data from file
func DataLoad(data interface{}, filename string) error { // 信息读取
	raw, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	byteGarble(raw)
	buffer := bytes.NewBuffer(raw)
	dec := gob.NewDecoder(buffer)
	return dec.Decode(data)
}

// byteGrable grable in place
func byteGarble(raw []byte) { // 字节混淆
	length := len(raw)
	for i := 0; i < length; i++ {
		raw[i] ^= byte(i % 256)
	}
}
