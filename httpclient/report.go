package httpclient

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type htmlSymbol uint8

const (
	symbolJSON htmlSymbol = iota
	symbolString

	reportURI = "/Mobile/rsbulid/r_3_3_st_jkdk.aspx"
)

var (
	//ErrIncompleteForm the form is incomplete
	ErrIncompleteForm = errors.New("form: incomplete form")
)

var fields = [...]string{"__EVENTARGUMENT", "__VIEWSTATE", "__VIEWSTATEENCRYPTED", "__VIEWSTATEGENERATOR",
	"bdbz", "bjhm", "brcnnrss", "brjkqk", "brjkqkdm", "ck_brcnnrss", "cw", "czsj",
	"databcdel", "databcxs", "dcbz", "fjmf", "hjzd", "jjzt", "jkmys", "jkmysdm", "lszt",
	"mc", "msie", "ndbz", "pa", "pb", "pc", "pd", "pe", "pf", "pg", "pkey", "pkey4", "psrc",
	"pzd_lock", "pzd_lock2", "pzd_lock3", "pzd_lock4", "pzd_y", "qx2_d", "qx2_i", "qx2_r",
	"qx2_u", "qx_d", "qx_i", "qx_r", "qx_u", "sfjczgfx", "sfjczgfxdm", "sfzx", "sfzxdm",
	"smbz", "st_nd", "st_xq", "tbrq", "tkey", "tkey4", "twqk", "twqkdm", "tzrjkqk", "tzrjkqkdm",
	"uname", "xcmqk", "xcmqkdm", "xdm", "xh", "xm", "xqbz", "xs_bj", "xzbz",
}

var fixedFields = map[string]string{"__EVENTTARGET": "databc"}

// getFormDetail 获取打卡表单详细信息
func (c *punchClient) getFormDetail() (form map[string]string, err error) {
	var req *http.Request
	req, err = getWithContext(c.ctx, host+reportURI)
	if err != nil {
		return
	}

	var res *http.Response
	if res, err = c.httpClient.Do(req); err != nil {
		return
	}
	defer drainBody(res.Body)
	bufferReader := bufio.NewReader(res.Body)
	form = make(map[string]string, len(fields)+len(fixedFields))
	for _, key := range fields {
		form[key] = ""
	}

	var key, value string
	for {
		key, value, err = parseHTML(bufferReader, "<input")
		if err != nil {
			break
		}
		if _, ok := form[key]; ok {
			form[key] = value
		}
	}
	for key, value := range fixedFields {
		form[key] = value
	}
	if err == io.EOF {
		err = nil
	}
	if err != nil {
		err = fmt.Errorf("get form data failed, err: %w", err)
	}
	return
}

// postForm 提交打卡表单
func (c *punchClient) postForm(form map[string]string) error {
	value := make(url.Values, len(form))
	for key, val := range form {
		value.Set(key, val)
	}

	req, err := postFormWithContext(c.ctx,
		host+reportURI,
		value,
	)
	if err != nil {
		return err
	}

	var res *http.Response
	if res, err = c.httpClient.Do(req); err != nil {
		return err
	}
	defer drainBody(res.Body)

	if res.StatusCode != http.StatusOK {
		return errors.New("post failed, status: " + res.Status)
	}
	_, errorMsg, _ := parseHTML(bufio.NewReader(res.Body), `<input name="cw"`) // get the error message
	switch errorMsg {
	case "保存修改成功!", "增加记录成功!":
		// success
	case "信息填报不完整\r\n保存失败!":
		err = ErrIncompleteForm
	case "":
		err = errors.New("post failed")
	default:
		err = errors.New("post failed, err: " + errorMsg)
	}
	return err
}
