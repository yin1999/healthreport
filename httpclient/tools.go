package httpclient

import (
	"bufio"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
)

var generalHeaders = http.Header{
	"Accept":          []string{"*/*"},
	"Accept-Language": []string{"zh-CN,zh;q=0.9"},
	"Connection":      []string{"keep-alive"},
	"User-Agent":      []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/99.0.0.0 Safari/537.36"},
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
	req.Header = generalHeaders.Clone()
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req, err
}

func getWithContext(ctx context.Context, url string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx,
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		return nil, err
	}
	req.Header = generalHeaders.Clone()
	return req, err
}

func notRedirect(req *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}

func encryptAES(data, key string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}
	padLen := aes.BlockSize - len(data)%aes.BlockSize
	cipherText := make([]byte, 64+len(data)+padLen)
	randBytes(cipherText[:64])
	copy(cipherText[64:], data)
	if padLen > 0 { // pkcs7padding
		fillBytes(cipherText[64+len(data):], byte(padLen))
	}
	iv := make([]byte, aes.BlockSize)
	randBytes(iv)

	mode := cipher.NewCBCEncrypter(block, iv)

	mode.CryptBlocks(cipherText, cipherText)

	return base64.StdEncoding.EncodeToString(cipherText), nil
}

// randBytes generate random bytes
func randBytes(data []byte) {
	const fill = "ABCDEFGHJKMNPQRSTWXYZabcdefhijkmnprstwxyz2345678"
	rand.Read(data)
	for i := range data {
		data[i] = fill[data[i]%byte(len(fill))]
	}
}

// fillBytes fill dst with data
func fillBytes(dst []byte, data byte) {
	if len(dst) == 0 {
		return
	}
	bp := 1
	dst[0] = data
	for bp < len(dst) {
		copy(dst[bp:], dst[:bp])
		bp <<= 1
	}
}

// scanLine scan a line
func scanLine(reader *bufio.Reader) (string, error) {
	data, isPrefix, err := reader.ReadLine() // data is not a copy, use it carefully
	res := string(textproto.TrimBytes(data)) // copy the data to string(remove the leading and trailing space)
	for isPrefix {                           // discard the remaining runes in the line
		_, isPrefix, err = reader.ReadLine()
	}
	return res, err
}

// drainBody discard all the data from reader and then close the reader
func drainBody(body io.ReadCloser) {
	io.Copy(io.Discard, body)
	body.Close()
}
