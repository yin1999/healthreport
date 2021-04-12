package httpclient

import (
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/google/go-querystring/query"
)

var (
	generalHeaders = [...]header{
		{"Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
		{"Accept-Encoding", "gzip"},
		{"Accept-Language", "zh-CN,zh;q=0.9"},
		{"Connection", "keep-alive"},
		{"User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36"},
	}

	prefixArray = [...]string{"    var _selfFormWid", "    var _userId", "            dataDetail", "            fillDetail"}
	symbolArray = [...]htmlSymbol{symbolString, symbolString, symbolJSON, symbolJSON}

	timeZone = time.Local // 打卡时区设置，默认为local
)

// LoginConfirm 验证账号密码
func LoginConfirm(ctx context.Context, account [2]string, timeout time.Duration) error {
	var cc context.CancelFunc
	ctx, cc = context.WithTimeout(ctx, timeout)
	_, err := login(ctx, account)
	cc()
	return parseURLError(err)
}

// Punch 打卡
func Punch(ctx context.Context, account [2]string, timeout time.Duration) error {
	var cc context.CancelFunc
	ctx, cc = context.WithTimeout(ctx, timeout)
	defer cc()

	cookies, err := login(ctx, account) // 登录，获取cookie
	if err != nil {
		return parseURLError(err)
	}

	cookies, err = getFormSessionID(ctx, cookies) // 获取打卡系统的cookie
	if err != nil {
		return parseURLError(err)
	}

	var (
		form   *HealthForm
		params *QueryParam
	)
	form, params, err = getFormDetail(ctx, cookies) // 获取打卡列表信息
	if err != nil {
		return parseURLError(err)
	}

	err = postForm(ctx, form, params, cookies) // 提交表单

	return parseURLError(err)
}

// SetTimeZone 设置时区
// 默认时区为 time.Local
func SetTimeZone(tz *time.Location) {
	if tz != nil {
		atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&timeZone)), unsafe.Pointer(tz))
	}
}

// getFormDetail 获取打卡表单详细信息
func getFormDetail(ctx context.Context, cookies []*http.Cookie) (*HealthForm, *QueryParam, error) {
	req, err := http.NewRequestWithContext(ctx,
		http.MethodGet,
		"http://form.hhu.edu.cn/pdc/formDesignApi/S/gUTwwojq",
		http.NoBody)

	if err != nil {
		return nil, nil, err
	}

	setGeneralHeader(req)
	setCookies(req, cookies)

	var res *http.Response

	if res, err = http.DefaultClient.Do(req); err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()

	var reader *gzip.Reader

	if reader, err = gzip.NewReader(res.Body); err != nil { // gzip数据解压
		return nil, nil, err
	}
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() && scanner.Text() != "<script type=\"text/javascript\">" {
	}

	var (
		line    string
		resData [4][]byte // wid, userId, personalInfo, healthFormData
		index   int
	)

	for index = 0; scanner.Scan() && index != 4; {
		line = scanner.Text()
		if strings.HasPrefix(line, prefixArray[index]) {
			if resData[index], err = parseData(line, symbolArray[index]); err != nil {
				return nil, nil, err
			}
			index++
		}
	}

	if index != 4 {
		return nil, nil, ErrCannotParseData
	}

	person := &PersonalInfo{}

	if err = json.Unmarshal(resData[indexPersonalInfo], person); err != nil {
		return nil, nil, err
	}

	form := &HealthForm{}

	if err = json.Unmarshal(resData[indexHealthFormData], form); err != nil {
		return nil, nil, err
	}

	form.Name = person.Name
	form.ID = person.ID
	form.DataTime = time.Now().In(timeZone).Format("2006/01/02") // 表单中增加打卡日期

	params := &QueryParam{
		wid:    string(resData[indexWID]),
		userID: string(resData[indexUserId]),
	}

	return form, params, nil
}

// getFormSessionID 获取打卡系统的SessionID
func getFormSessionID(ctx context.Context, cookies []*http.Cookie) ([]*http.Cookie, error) {
	req, err := http.NewRequestWithContext(ctx,
		http.MethodGet,
		"http://form.hhu.edu.cn/pdc/form/list",
		http.NoBody)

	if err != nil {
		return nil, err
	}

	setGeneralHeader(req)
	setCookies(req, cookies)

	var res *http.Response
	if res, err = http.DefaultClient.Do(req); err != nil {
		return nil, err
	}
	res.Body.Close()

	if cookie := getCookie(res.Cookies(), "JSESSIONID"); cookie != nil {
		return []*http.Cookie{cookie}, nil
	}

	return nil, CookieNotFoundErr{"JSESSIONID"}
}

// login 登录系统
func login(ctx context.Context, account [2]string) ([]*http.Cookie, error) {
	data := url.Values{}
	data.Set("IDToken1", account[0])
	data.Set("IDToken2", account[1])

	req, err := http.NewRequestWithContext(ctx,
		http.MethodPost,
		"http://ids.hhu.edu.cn/amserver/UI/Login",
		strings.NewReader(data.Encode()))

	if err != nil {
		return nil, err
	}

	setGeneralHeader(req)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{
		CheckRedirect: getResponseN(1),
	}

	var res *http.Response
	if res, err = client.Do(req); err != nil {
		return nil, err
	}
	res.Body.Close()

	if cookie := getCookie(res.Cookies(), "iPlanetDirectoryPro"); cookie != nil {
		return []*http.Cookie{cookie}, nil
	}

	return nil, CookieNotFoundErr{"iPlanetDirectoryPro"}
}

// parseURLError 解析URL错误
func parseURLError(err error) error {
	if err == nil {
		return err
	}
	if v, ok := err.(*url.Error); ok {
		return v.Err
	}
	return err
}

// postForm 提交打卡表单
func postForm(ctx context.Context, form *HealthForm, params *QueryParam, cookies []*http.Cookie) error {
	value, err := query.Values(form)
	if err != nil {
		return err
	}

	buf := value.Encode()

	req, err := http.NewRequestWithContext(ctx,
		http.MethodPost,
		"http://form.hhu.edu.cn/pdc/formDesignApi/dataFormSave",
		strings.NewReader(buf))

	if err != nil {
		return err
	}

	setGeneralHeader(req)
	setCookies(req, cookies)

	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	queryValue := url.Values{}
	queryValue.Set("wid", params.wid)
	queryValue.Set("userId", params.userID)

	req.URL.RawQuery = queryValue.Encode()

	var res *http.Response
	if res, err = http.DefaultClient.Do(req); err != nil {
		return err
	}
	res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("post failed, status code: %d", res.StatusCode)
	}

	return nil
}

// getCookie get Cookie by name
func getCookie(cookies []*http.Cookie, name string) *http.Cookie {
	for i := range cookies {
		if cookies[i].Name == name {
			return cookies[i]
		}
	}
	return nil
}

func getResponseN(n int) func(req *http.Request, via []*http.Request) error {
	return func(req *http.Request, via []*http.Request) error {
		if len(via) == n {
			return http.ErrUseLastResponse
		}
		return nil
	}
}

func parseData(data string, symbol htmlSymbol) (res []byte, err error) {
	switch symbol {
	case symbolJSON:
		res, err = getSlice(data, '{', '}')
	case symbolString:
		res, err = getSlice(data, '\'', '\'')
		res = res[1 : len(res)-1]
	default:
		err = ErrInvalidSymbol
	}

	return
}

func getSlice(data string, startSymbol, endSymbol byte) (res []byte, err error) {
	start := strings.IndexByte(data, startSymbol)
	if start == -1 {
		return nil, ErrCannotParseData
	}

	length := strings.IndexByte(data[start+1:], endSymbol)
	if length == -1 {
		return nil, ErrCannotParseData
	}

	res = make([]byte, length+2)
	copy(res, data[start:])

	return res, err
}

func setCookies(req *http.Request, cookies []*http.Cookie) {
	for i := range cookies {
		req.AddCookie(cookies[i])
	}
}

func setGeneralHeader(req *http.Request) {
	for i := range generalHeaders {
		req.Header.Set(generalHeaders[i].key, generalHeaders[i].value)
	}
}
