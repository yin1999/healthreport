package httpclient

import (
	"context"
	"net/http"
)

// QueryParam query param struct
type QueryParam struct {
	Wid    string `url:"wid"`
	UserID string `url:"userId"`
}

// HealthForm form for post
type HealthForm struct {
	DataTime        string `json:"DATETIME_CYCLE" url:"DATETIME_CYCLE"` // 填报时间
	StudentID       string `json:"XGH_336526" url:"XGH_336526"`         // 学号
	Name            string `json:"XM_1474" url:"XM_1474"`               // 姓名
	ID              string `json:"SFZJH_859173" url:"SFZJH_859173"`     // 身份证号
	College         string `json:"SELECT_941320" url:"SELECT_941320"`   // 学院
	Grade           string `json:"SELECT_459666" url:"SELECT_459666"`   // 年级
	Department      string `json:"SELECT_814855" url:"SELECT_814855"`   // 专业
	Class           string `json:"SELECT_525884" url:"SELECT_525884"`   // 班级
	Dormitory       string `json:"SELECT_125597" url:"SELECT_125597"`   // 宿舍楼
	DormitoryNumber string `json:"TEXT_950231" url:"TEXT_950231"`       // 宿舍号
	PhoneNumber     string `json:"TEXT_937296" url:"TEXT_937296"`       // 手机号码
	Temperature     string `json:"RADIO_6555" url:"RADIO_6555"`         // 体温状况
	InSchool        string `json:"RADIO_535015" url:"RADIO_535015"`     // 是否在校内
	SelfHealth      string `json:"RADIO_891359" url:"RADIO_891359"`     // 本人健康状况
	MateHealth      string `json:"RADIO_372002" url:"RADIO_372002"`     // 同住人健康状况
	HighRisk        string `json:"RADIO_618691" url:"RADIO_618691"`     // 中高风险地区旅居史、接触中高风险地区人员
}

// CookieNotFoundErr error interface for Cookies
type CookieNotFoundErr struct {
	cookie string
}

func (t CookieNotFoundErr) Error() string {
	return "http: can't find cookie: " + t.cookie
}

type punchClient struct {
	ctx        context.Context
	httpClient *http.Client
	jar        customCookieJar
}

// Account account info for login
type Account struct {
	Username string
	Password string
}

// Name get the name of the account
func (a Account) Name() string {
	return a.Username
}
