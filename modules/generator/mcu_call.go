package generator

// McuCallDetail detail info such as number and params
type McuCallDetail struct {
	Number string `json:"number"`
}

// McuCall Audit Record Definition for mcu call
type McuCall struct {
	Type       string        `json:"type"`
	GUID       string        `json:"guid"`
	confGUID   string        `json:"conf_guid"`
	CallDetail McuCallDetail `json:"call"`
	StartTime  string        `json:"start_time"`
	EndTime    string        `json:"end_time"`
	Duration   int           `json:"duration"`
	ErrorCode  int           `json:"error_code"`
	ErrorInfo  string        `json:"error_info"`
}
