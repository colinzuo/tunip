package generator

// ModuleName for log
const (
	ModuleName string = "Generator"
)

// Config config for Generator
type Config struct {
	GenMcuConf bool `json:"gen_mcu_conf"`
	McuConfNum int  `json:"mcu_conf_num"`

	GenMcuCall    bool `json:"gen_mcu_call"`
	McuCallNumMin int  `json:"mcu_call_num_min"`
	McuCallNumMax int  `json:"mcu_call_num_max"`
}
