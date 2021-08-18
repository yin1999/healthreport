package httpclient

import (
	"bufio"
	"context"
	"encoding/xml"
	"errors"
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
func login(ctx context.Context, account [2]string) (jar customCookieJar, err error) {
	const loginURL = "https://authserver.hhu.edu.cn/authserver/login"
	var req *http.Request
	req, err = getWithContext(ctx, loginURL)
	if err != nil {
		return
	}

	jar = newCookieJar()
	client := &http.Client{Jar: jar}

	var res *http.Response
	if res, err = client.Do(req); err != nil {
		return
	}
	defer res.Body.Close()
	f := &loginForm{}

	{
		bufferReader := bufio.NewReader(res.Body)
		var line string
		const inputElement = "<input type=\"hidden\""
		for !strings.HasPrefix(line, inputElement) {
			line, err = scanLine(bufferReader)
			if err != nil {
				return
			}
		}

		var filler *structFiller
		if filler, err = newFiller(f, "fill"); err != nil {
			return
		}
		var v *elementInput
		for {
			v, err = elementParse(line)
			if err != nil {
				return
			}
			filler.fill(v.Key, v.Value)
			line, _ = scanLine(bufferReader)
			if !strings.HasPrefix(line, inputElement) {
				break
			}
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

	req, err = postFormWithContext(ctx, loginURL, value)
	if err != nil {
		return
	}

	client.CheckRedirect = getResponseN(1)
	if res, err = client.Do(req); err != nil {
		return
	}
	res.Body.Close()

	if jar.GetCookieByName("iPlanetDirectoryPro") == nil {
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
