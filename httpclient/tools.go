package httpclient

import (
	"bufio"
	"context"
	"encoding/xml"
	"io"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
)

const host = "http://smst.hhu.edu.cn"

var generalHeaders = http.Header{
	"Accept":          []string{"*/*"},
	"Accept-Language": []string{"zh-CN,zh;q=0.9"},
	"Connection":      []string{"keep-alive"},
	"User-Agent":      []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.0.0 Safari/537.36"},
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

func parseHTML(bufferReader *bufio.Reader, prefix string) (key string, value string, err error) {
	var line string
	var v *elementInput
	for {
		if line, err = scanLine(bufferReader); err != nil {
			break
		}
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		v, err = elementParse(line)
		if err != nil {
			break
		}
		return v.Key, v.Value, err
	}
	return
}

type elementInput struct {
	Key   string `xml:"name,attr"`
	Value string `xml:"value,attr"`
	ID    string `xml:"id,attr"`
}

func elementParse(v string) (*elementInput, error) {
	if len(v) < 2 {
		return nil, &xml.SyntaxError{Msg: "error format", Line: 1}
	}
	out := &elementInput{}
	data := []byte(v)
	if data[len(data)-2] != '/' && strings.Index(v, "</") == -1 {
		data = append(data[:len(data)-1], '/', '>')
	}
	err := xml.Unmarshal(data, out)
	if err != nil {
		return nil, err
	}
	if out.Key == "" {
		out.Key = out.ID
	}
	return out, err
}

// fillMap fill the map with the key and value when `use` returns true.
// If use is nil, all the key and value will be filled
func fillMap(reader io.Reader, v url.Values, use func(string) bool) error {
	if use == nil {
		use = func(string) bool { return true }
	}
	bufferReader := bufio.NewReader(reader)
	for {
		key, value, err := parseHTML(bufferReader, "<input ")
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return err
		}
		if use(key) {
			v.Set(key, value)
		}
	}
}
