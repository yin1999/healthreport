package httpclient

import (
	"bufio"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/google/go-querystring/query"
)

type loginForm struct {
	Username   string `url:"username"`
	Password   string `url:"password"`
	Session    string `fill:"lt" url:"lt"`
	Method     string `fill:"dllt" url:"dllt"`
	Excution   string `fill:"execution" url:"execution"`
	Event      string `fill:"_eventId" url:"_eventId"`
	Show       string `fill:"rmShown" url:"rmShown"`
	EncryptKey string `fill:"pwdDefaultEncryptSalt" url:"-"`
}

// login 登录系统
func login(ctx context.Context, account [2]string) (cookies []*http.Cookie, err error) {
	const loginURL = "http://authserver.hhu.edu.cn/authserver/login"
	var req *http.Request
	req, err = getWithContext(ctx, loginURL)
	if err != nil {
		return
	}
	setGeneralHeader(req)

	var res *http.Response
	if res, err = http.DefaultClient.Do(req); err != nil {
		return
	}
	var reader io.ReadCloser
	if reader, err = responseReader(res); err != nil {
		return
	}
	defer reader.Close()
	f := &loginForm{}
	{
		bufferReader := bufio.NewReader(res.Body)
		var line string
		for err == nil && !strings.HasPrefix(line, "<input type=\"hidden\"") {
			line, err = scanLine(bufferReader)
		}
		var v *elementInput
		filler := newFiller(f, "fill")
		for ; strings.HasPrefix(line, "<input type=\"hidden\""); line, _ = scanLine(bufferReader) {
			v, err = elementParse(line)
			if err != nil {
				return
			}
			filler.fill(v.Key, v.Value)
		}
	}
	f.Username = account[0]
	f.Password, err = encryptAES(account[1], f.EncryptKey)
	if err != nil {
		return
	}

	var value url.Values
	if value, err = query.Values(f); err != nil {
		return
	}

	req, err = postWithContext(ctx, loginURL, value)
	if err != nil {
		return
	}
	setCookies(req, res.Cookies())

	client := http.Client{CheckRedirect: getResponseN(1)}
	res, err = client.Do(req)
	if err != nil {
		return
	}
	res.Body.Close()

	if cookies = getCookie(res.Cookies(), []string{"iPlanetDirectoryPro"}); cookies == nil {
		err = CookieNotFoundErr{"iPlanetDirectoryPro"}
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
	if data[len(data)-2] != '/' {
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

// newFiller default tag: fill
func newFiller(item interface{}, tag string) *structFiller {
	v := reflect.ValueOf(item).Elem()
	if !v.CanAddr() {
		panic("reflect: item must be a pointer")
	}
	if tag == "" {
		tag = "fill"
	}
	findTagName := func(t reflect.StructTag) (string, error) {
		if tn, ok := t.Lookup(tag); ok {
			return strings.Split(tn, ",")[0], nil
		}
		return "", errors.New("reflect: not define a" + tag + "tag")
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
	return s
}

func (s *structFiller) fill(key string, value interface{}) error {
	fieldNum, ok := s.m[key]
	if !ok {
		return fmt.Errorf("reflect: field %s not exists", key)
	}
	s.v.Field(fieldNum).Set(reflect.ValueOf(value))
	return nil
}
