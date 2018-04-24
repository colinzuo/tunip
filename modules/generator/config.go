package generator

// ModuleName for log
const (
	ModuleName string = "Generator"
)

// McuConfConfig config for mcu conf
type McuConfConfig struct {
	Num         int    `json:"num"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	NumberMin   int    `json:"number_min"`
	NumberMax   int    `json:"number_max"`
	DurationMin int    `json:"duration_min"`
	DurationMax int    `json:"duration_max"`
}

// McuCallConfig config for mcu call
type McuCallConfig struct {
	NumMin      int `json:"num_min"`
	Num25       int `json:"num_25"`
	Num50       int `json:"num_50"`
	Num75       int `json:"num_75"`
	NumMax      int `json:"num_max"`
	DurationMin int `json:"duration_min"`
	DurationMax int `json:"duration_max"`
}

// Config config for Generator
type Config struct {
	ServerAddr    string        `json:"server_addr"`
	GenMcuConf    bool          `json:"gen_mcu_conf"`
	McuConfConfig McuConfConfig `json:"mcu_conf_config"`

	GenMcuCall    bool          `json:"gen_mcu_call"`
	McuCallConfig McuCallConfig `json:"mcu_call_config"`
}
