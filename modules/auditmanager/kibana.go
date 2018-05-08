package auditmanager

import (
	"github.com/colinzuo/tunip/thirdparty/elastic/beats/libbeat/dashboards"
)

func (m manager) loadDashboard() error {
	m.logger.Infof("Enter loadDashboard")
	defer m.logger.Infof("Leave loadDashboard")

	homePath, _ := m.config.BeatConfig.String("path.home", -1)

	err := dashboards.ImportDashboards(homePath, m.config.SetupConfig.Kibana, m.config.SetupConfig.Dashboards)

	if err != nil {
		m.logger.Errorf("Failed to load dashboard: %s", err)
	}

	return err
}
