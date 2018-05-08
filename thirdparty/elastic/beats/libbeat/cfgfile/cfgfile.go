package cfgfile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/colinzuo/tunip/thirdparty/elastic/beats/libbeat/common"
)

var (
	defaults = mustNewConfigFrom(map[string]interface{}{
		"path": map[string]interface{}{
			"home":   ".", // to be initialized by beat
			"config": "${path.home}",
		},
	})
)

func mustNewConfigFrom(from interface{}) *common.Config {
	cfg, err := common.NewConfigFrom(from)
	if err != nil {
		panic(err)
	}
	return cfg
}

// HandleFlags update default "path.home"
func HandleFlags() error {
	// default for the home path is the binary location
	home, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return fmt.Errorf("The absolute path to %s could not be obtained. %v",
			os.Args[0], err)
	}

	defaults.SetString("path.home", -1, home)

	return nil
}

// Load reads the configuration from a YAML file structure
func Load(path string) (*common.Config, error) {
	var config *common.Config
	var err error

	if path == "" {
		return nil, errors.New("path shouldn't be empty")
	}

	config, err = common.LoadFile(path)

	if err != nil {
		return nil, err
	}

	config, err = common.MergeConfigs(
		defaults,
		config,
	)
	if err != nil {
		return nil, err
	}

	config.PrintDebugf("Complete configuration loaded:")
	return config, nil
}
