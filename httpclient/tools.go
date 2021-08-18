package httpclient

import (
	"bufio"
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
)

var generalHeaders = [...]header{
	{"Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"},
	{"Accept-Language", "zh-CN,zh;q=0.9"},
	{"Connection", "keep-alive"},
	{"User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
}

func postFormWithContext(ctx context.Context, url string, data url.Values) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx,
		http.MethodPost,
		url,
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return nil, err
	}
	setGeneralHeader(req)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req, err
}

func getWithContext(ctx context.Context, url string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx,
		http.MethodGet,
		url,
		http.NoBody,
	)
	if err != nil {
		return nil, err
	}
	setGeneralHeader(req)
	return req, err
}

var isAsciiSpace = [256]bool{'\t': true, '\n': true, '\v': true, '\f': true, '\r': true, ' ': true}

func trimSuffixSpace(data []byte) []byte {
	start := 0
	for start < len(data) && isAsciiSpace[data[start]] {
		start++
	}
	return data[start:]
}

func getResponseN(n int) func(req *http.Request, via []*http.Request) error {
	return func(req *http.Request, via []*http.Request) error {
		if len(via) == n {
			return http.ErrUseLastResponse
		}
		return nil
	}
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padLen := blockSize - len(data)%blockSize
	padding := bytes.Repeat([]byte{byte(padLen)}, padLen)
	return append(data, padding...)
}

func encryptAES(data, key string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}
	plainText := pkcs7Pad([]byte(data), aes.BlockSize)
	cipherText := make([]byte, 64+len(plainText))
	iv := make([]byte, 16)
	randBytes(cipherText[:64])
	randBytes(iv)

	copy(cipherText[64:], plainText)

	mode := cipher.NewCBCEncrypter(block, iv)

	mode.CryptBlocks(cipherText, cipherText)

	return string(base64.StdEncoding.EncodeToString(cipherText)), nil
}

// randBytes generate random bytes
func randBytes(data []byte) {
	const fill = "ABCDEFGHJKMNPQRSTWXYZabcdefhijkmnprstwxyz2345678"
	const length = int32(len(fill))
	for i := range data {
		data[i] = fill[rand.Int31()%length]
	}
}

// scanLine scan a line
func scanLine(reader *bufio.Reader) (string, error) {
	data, isPrefix, err := reader.ReadLine() // data is not a copy, use it carefully
	res := string(trimSuffixSpace(data))     // copy the data to string(remove the leading space)
	for isPrefix {                           // discard the remaining runes in the line
		_, isPrefix, err = reader.ReadLine()
	}
	return res, err
}

func setGeneralHeader(req *http.Request) {
	for i := range generalHeaders {
		req.Header.Set(generalHeaders[i].key, generalHeaders[i].value)
	}
}
