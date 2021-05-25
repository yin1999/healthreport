package httpclient

type header struct {
	key   string
	value string
}

// QueryParam query param struct
type QueryParam struct {
	Wid    string `url:"wid"`
	UserID string `url:"userId"`
}

// HealthForm form for post
type HealthForm struct {
	DataTime                string `json:"DATETIME_CYCLE" url:"DATETIME_CYCLE"` // 填报时间
	StudentID               string `json:"XGH_336526" url:"XGH_336526"`         // 学号
	Name                    string `json:"XM_1474" url:"XM_1474"`               // 姓名
	ID                      string `json:"SFZJH_859173" url:"SFZJH_859173"`     // 身份证号
	College                 string `json:"SELECT_941320" url:"SELECT_941320"`   // 学院
	Grade                   string `json:"SELECT_459666" url:"SELECT_459666"`   // 年级
	Department              string `json:"SELECT_814855" url:"SELECT_814855"`   // 专业
	Class                   string `json:"SELECT_525884" url:"SELECT_525884"`   // 班级
	Dormitory               string `json:"SELECT_125597" url:"SELECT_125597"`   // 宿舍楼
	DormitoryNumber         string `json:"TEXT_950231" url:"TEXT_950231"`       // 宿舍号
	PhoneNumber             string `json:"TEXT_937296" url:"TEXT_937296"`       // 手机号码
	MorningTemperature      string `json:"RADIO_853789" url:"RADIO_853789"`     // 上午体温
	AfternoonTemperature    string `json:"RADIO_43840" url:"RADIO_43840"`       // 下午体温
	HealthCondition         string `json:"RADIO_579935" url:"RADIO_579935"`     // 健康情况
	InSchool                string `json:"RADIO_138407" url:"RADIO_138407"`     // 是否返校
	InCaseOfCorona          string `json:"RADIO_546905" url:"RADIO_546905"`     // 新冠肺炎病例
	CloseContact1           string `json:"RADIO_314799" url:"RADIO_314799"`     // 新冠肺炎病例密切接触者
	SeverelyAffectedAreas   string `json:"RADIO_209256" url:"RADIO_209256"`     // 疫情严重地区
	CloseContact2           string `json:"RADIO_836972" url:"RADIO_836972"`     // 发热病例密切接触
	HistoryOfOverseasTravel string `json:"RADIO_302717" url:"RADIO_302717"`     // 海外居旅史
	CloseContact3           string `json:"RADIO_701131" url:"RADIO_701131"`     // 境外人员密切接触
	Isolation               string `json:"RADIO_438985" url:"RADIO_438985"`     // 隔离状态
	Domestic                string `json:"RADIO_467360" url:"RADIO_467360"`     // 是否在国内
	Address                 string `json:"PICKER_956186" url:"PICKER_956186"`   // 住址
	Country                 string `json:"TEXT_434598" url:"TEXT_434598"`       // 国家
	City                    string `json:"TEXT_515297" url:"TEXT_515297"`       // 城市
	School                  string `json:"TEXT_752063" url:"TEXT_752063"`       // 学校
}

// CookieNotFoundErr error interface for Cookies
type CookieNotFoundErr struct {
	cookie string
}

func (t CookieNotFoundErr) Error() string {
	return "http: can't find cookie: " + t.cookie
}
