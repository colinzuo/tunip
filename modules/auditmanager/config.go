package auditmanager

import (
	"github.com/colinzuo/tunip/logp"
	"github.com/colinzuo/tunip/thirdparty/elastic/beats/libbeat/cfgfile"
	"github.com/colinzuo/tunip/thirdparty/elastic/beats/libbeat/common"
	"github.com/spf13/viper"
)

// ModuleName for log
const (
	ModuleName string = "AuditManager"
)

// Config config for audit manager
type Config struct {
	MaxWorker      int    `json:"maxworker"`
	WebPort        int    `json:"webport"`
	ReqTimeout     int    `json:"reqtimeout"`
	BatchSize      int    `json:"batchsize"`
	BatchTimeout   int    `json:"batchtimeout"`
	EsServerAddr   string `json:"esserveraddr"`
	Setup          bool   `json:"setup"`
	BeatConfigPath string `json:"beatconfigpath"`
	BeatConfig     *common.Config
	SetupConfig    SetupConfig
}

// SetupConfig elastic stack 'setup' configurations
type SetupConfig struct {
	Dashboards *common.Config `config:"setup.dashboards"`
	Template   *common.Config `config:"setup.template"`
	Kibana     *common.Config `config:"setup.kibana"`
}

// API response error code
const (
	ErrCodeOk                = 0
	ErrCodeFailedToReadBody  = 10000
	ErrCodeFailedToParseBody = 10001
	ErrCodeTimeout           = 10002
	ErrCodeIndex             = 10003
	ErrCodeGeneral           = 10004
	ErrCodeBadFormat         = 10005
	ErrCodeUnexpected        = 10006
)

// API response error info
const (
	ErrInfoOk                = "RESULT_OK"
	ErrInfoFailedToReadBody  = "ErrInfoFailedToReadBody"
	ErrInfoFailedToParseBody = "ErrInfoFailedToParseBody"
	ErrInfoTimeout           = "ErrInfoTimeout"
	ErrInfoIndex             = "ErrInfoIndex"
	ErrInfoGeneral           = "ErrInfoGeneral"
	ErrInfoBadFormat         = "ErrInfoBadFormat"
	ErrInfoUnexpected        = "ErrInfoUnexpected"
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
	MaxWorker:      8,
	WebPort:        8080,
	ReqTimeout:     3000,
	BatchSize:      1000,
	BatchTimeout:   300,
	Setup:          false,
	BeatConfigPath: "",
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

	if len(config.BeatConfigPath) > 0 {
		err = cfgfile.HandleFlags()

		if err != nil {
			logger.Panic("HandleFlags failed, %s", err)
		}

		config.BeatConfig, err = cfgfile.Load(config.BeatConfigPath)

		if err != nil {
			logger.Panic("read BeatConfig %s failed, %s", config.BeatConfigPath, err)
		}

		err = config.BeatConfig.Unpack(&config.SetupConfig)

		if err != nil {
			logger.Panic("Unpack SetupConfig failed, %s", err)
		}
	}

	return config, nil
}
