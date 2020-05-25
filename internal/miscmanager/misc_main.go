package miscmanager

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"time"

	"github.com/colinzuo/tunip/pkg/logp"
)

// Manager def
type Manager struct {
	logger *logp.Logger
	config *Config

	dispatchChan   chan WorkerRequest
	freeWorkerChan chan chan interface{}
	doneChan       chan bool

	timeLongForm string
}

// ParseConfig parse config
func ParseConfig(configPath string) (*Config, error) {
	logger := logp.NewLogger(ModuleName)
	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		logger.Errorf("read config %s failed, %s", configPath, err)
		return nil, err
	}

	var config Config
	err = json.Unmarshal(content, &config)
	if err != nil {
		logger.Errorf("parse config %s failed, %s", configPath, err)
		return nil, err
	}
	return &config, nil
}

// Run function to start
func Run(configPath string) {
	logger := logp.NewLogger(ModuleName)
	var config *Config
	config, err := ParseConfig(configPath)
	if err != nil {
		logger.Panicf("ParseConfig: failed with %s", err)
	}
	logger.Infof("config %s, content: %+v", configPath, config)

	manager := Manager{logger: logger, config: config,
		timeLongForm: "2006-01-02T15:04:05.000-07:00"}
	manager.Work()
}

// Work work according to config
func (m *Manager) Work() {
	logger := m.logger
	config := m.config

	jsonRsp, _ := json.MarshalIndent(config, "", "    ")

	logger.Infof("Enter with config: %s", string(jsonRsp))

	rand.Seed((int64)(time.Now().Second()))

	if config.PerfTest != nil && config.PerfTest.Enabled {
		m.perfTest()
	}

	if config.ViperTest != nil && config.ViperTest.Enabled {
		m.viperTest()
	}
}
