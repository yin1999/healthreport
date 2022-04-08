package httpclient

import (
	"bufio"
	"context"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"net/textproto"
	"net/url"
	"reflect"
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

type structFiller struct {
	m map[string]int
	v reflect.Value
}

// newFiller default tag: fill.
// The item must be a pointer
func newFiller(item interface{}, tag string) (*structFiller, error) {
	v := reflect.ValueOf(item).Elem()
	if !v.CanAddr() {
		return nil, errors.New("reflect: item must be a pointer")
	}
	if tag == "" {
		tag = "fill"
	}
	findTagName := func(t reflect.StructTag) (string, error) {
		if tn, ok := t.Lookup(tag); ok && len(tn) > 0 {
			return strings.Split(tn, ",")[0], nil
		}
		return "", errors.New("skip")
	}
	s := &structFiller{
		m: make(map[string]int),
		v: v,
	}
	for i := 0; i < v.NumField(); i++ {
		typeField := v.Type().Field(i)
		name, err := findTagName(typeField.Tag)
		if err != nil {
			continue
		}
		s.m[name] = i
	}
	return s, nil
}

func (s *structFiller) fill(key string, value interface{}) error {
	fieldNum, ok := s.m[key]
	if !ok {
		return errors.New("reflect: field <" + key + "> not exists")
	}
	s.v.Field(fieldNum).Set(reflect.ValueOf(value))
	return nil
}
