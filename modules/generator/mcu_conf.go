package generator

// McuConfDetail detail info such as number and params
type McuConfDetail struct {
	Number string `json:"number"`
}

// McuConf Audit Record Definition for mcu conf
type McuConf struct {
	Type       string        `json:"type"`
	GUID       string        `json:"guid"`
	ConfDetail McuConfDetail `json:"conf"`
	StartTime  string        `json:"start_time"`
	EndTime    string        `json:"end_time"`
	Duration   int           `json:"duration"`
	ErrorCode  int           `json:"error_code"`
	ErrorInfo  string        `json:"error_info"`
}
