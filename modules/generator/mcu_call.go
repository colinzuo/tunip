package generator

// McuCall Audit Record Definition for mcu call
type McuCall struct {
	Type       string `json:"type"`
	GUID       string `json:"guid"`
	ConfGUID   string `json:"conf_guid"`
	ConfNumber string `json:"conf_number"`
	StartTime  string `json:"start_time"`
	EndTime    string `json:"end_time"`
	Duration   int    `json:"duration"`
	ErrorCode  int    `json:"error_code"`
	ErrorInfo  string `json:"error_info"`
}
