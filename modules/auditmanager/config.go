package auditmanager

import (
	"github.com/colinzuo/tunip/logp"
	"github.com/spf13/viper"
)

// ModuleName for log
const (
	ModuleName string = "AuditManager"
)

// Config config for audit manager
type Config struct {
	MaxWorker    int    `json:"maxworker"`
	WebPort      int    `json:"webport"`
	ReqTimeout   int    `json:"reqtimeout"`
	BatchSize    int    `json:"batchsize"`
	BatchTimeout int    `json:"batchtimeout"`
	EsServerAddr string `json:"esserveraddr"`
}

// API response error code
const (
	ErrCodeOk                = 0
	ErrCodeFailedToReadBody  = 10000
	ErrCodeFailedToParseBody = 10001
	ErrCodeTimeout           = 10002
	ErrCodeIndex             = 10003
	ErrCodeGeneral           = 10004
)

// API response error info
const (
	ErrInfoOk                = "RESULT_OK"
	ErrInfoFailedToReadBody  = "ErrInfoFailedToReadBody"
	ErrInfoFailedToParseBody = "ErrInfoFailedToParseBody"
	ErrInfoTimeout           = "ErrInfoTimeout"
	ErrInfoIndex             = "ErrInfoIndex"
	ErrInfoGeneral           = "ErrInfoGeneral"
)

// API response key
const (
	KeyErrCode = "err_code"
	KeyErrInfo = "err_info"
)

// Content Types
const (
	ContentTypeJSON = "application/json"
)

var defaultConfig = Config{
	MaxWorker:    8,
	WebPort:      8080,
	ReqTimeout:   3000,
	BatchSize:    1000,
	BatchTimeout: 300,
}

// DefaultConfig returns the default config options.
func DefaultConfig() Config {
	return defaultConfig
}

func initConfig() (Config, error) {
	logger := logp.NewLogger(ModuleName)

	config := DefaultConfig()
	err := viper.UnmarshalKey("audit_manager", &config)

	if err != nil {
		logger.Errorf("initConfig: Unmarshal failed with %s", err)
	}

	if config.WebPort <= 0 {
		logger.Panic("initConfig: WebPort should be larger than 0")
	}

	if config.ReqTimeout <= 100 {
		logger.Panic("initConfig: ReqTimeout should be larger than 100")
	}

	if config.BatchSize <= 100 {
		logger.Panic("initConfig: BatchSize should be larger than 100")
	}

	if config.BatchTimeout <= 50 {
		logger.Panic("initConfig: BatchTimeout should be larger than 50")
	}

	return config, nil
}
