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
	MaxWebListener int `json:"maxweblistener"`
	MaxWorker      int `json:"maxworker"`
	WebPort        int `json:"webport"`
}

// API response error code
const (
	ErrCodeOk               = 0
	ErrCodeFailedToReadBody = 10000
)

// API response error info
const (
	ErrInfoOk               = "RESULT_OK"
	ErrInfoFailedToReadBody = "ErrInfoFailedToReadBody"
)

// API response key
const (
	KeyErrCode = "err_code"
	KeyErrInfo = "err_info"
)

var defaultConfig = Config{
	MaxWebListener: 2000,
	MaxWorker:      8,
	WebPort:        8080,
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

	if config.WebPort < 0 {
		logger.Panic("initConfig: invalid WebPort configuration")
	}

	return config, nil
}
