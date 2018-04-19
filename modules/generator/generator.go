package generator

import (
	"encoding/json"
	"io/ioutil"

	"github.com/colinzuo/tunip/logp"
)

// ParseGenConfig parse generator config
func ParseGenConfig(genConfig string) (*Config, error) {
	logger := logp.NewLogger(ModuleName)
	content, err := ioutil.ReadFile(genConfig)
	if err != nil {
		logger.Errorf("read genConfig %s failed, %s", genConfig, err)
		return nil, err
	}

	var config Config
	err = json.Unmarshal(content, &config)
	if err != nil {
		logger.Errorf("parse genConfig %s failed, %s", genConfig, err)
		return nil, err
	}
	return &config, nil
}

// Generate generate fake data
func Generate(genConfig string) error {
	logger := logp.NewLogger(ModuleName)
	var config *Config
	config, err := ParseGenConfig(genConfig)
	if err != nil {
		logger.Panicf("ParseGenConfig: failed with %s", err)
	}
	logger.Infof("genConfig %s, content: %v", genConfig, config)
	return nil
}
