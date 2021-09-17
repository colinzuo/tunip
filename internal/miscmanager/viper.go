package miscmanager

import (
	"github.com/spf13/viper"
)

func (m *Manager) viperTest() {
	logger := m.logger
	config := m.config.ViperTest

	logger.Infof("Enter viperTest with config: %+v", config)
	defer logger.Info("Leave viperTest")

	value := viper.GetString(config.Key)

	logger.Infof("Key %s has value %s", config.Key, value)
}
