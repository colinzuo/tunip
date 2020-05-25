package miscmanager

// ModuleName for log
const (
	ModuleName string = "Misc"
)

// PerfTestConfig performance test config
type PerfTestConfig struct {
	Enabled   bool `json:"enabled"`
	MaxWorker int  `json:"maxworker"`
	Number    int  `json:"number"`
}

// ViperTestConfig viper test config
type ViperTestConfig struct {
	Enabled bool   `json:"enabled"`
	Key     string `json:"key"`
}

// Config config for Generator
type Config struct {
	ServerAddr string `json:"server_addr"`

	PerfTest  *PerfTestConfig  `json:"perf_test,omitempty"`
	ViperTest *ViperTestConfig `json:"viper_test,omitempty"`
}

// API response error code
const (
	ErrCodeOk                   = 0
	ErrCodeFailedToReadBody     = 10000
	ErrCodeFailedToParseBody    = 10001
	ErrCodeTimeout              = 10002
	ErrCodeGeneral              = 10003
	ErrCodeBadFormat            = 10004
	ErrCodeUnexpected           = 10005
	ErrCodeHTTPErr              = 10006
	ErrCodeFailedToParseRspBody = 10007
)

// API response error info
const (
	ErrMsgOk                   = "RESULT_OK"
	ErrMsgFailedToReadBody     = "ErrMsgFailedToReadBody"
	ErrMsgFailedToParseBody    = "ErrMsgFailedToParseBody"
	ErrMsgTimeout              = "ErrMsgTimeout"
	ErrMsgGeneral              = "ErrMsgGeneral"
	ErrMsgBadFormat            = "ErrMsgBadFormat"
	ErrMsgUnexpected           = "ErrMsgUnexpected"
	ErrMsgHTTPErr              = "ErrMsgHTTPErr"
	ErrMsgFailedToParseRspBody = "ErrMsgFailedToParseRspBody"
)

// API response key
const (
	KeyErrCode = "err_code"
	KeyErrMsg  = "err_msg"
)

// Content Types
const (
	ContentTypeJSON = "application/json"
)

// Request Type
const (
	RequestSampleWorkerReq = "RequestSampleWorkerReq"
)
