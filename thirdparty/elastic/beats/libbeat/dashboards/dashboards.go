package dashboards

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/colinzuo/tunip/thirdparty/elastic/beats/libbeat/common"
)

type importMethod uint8

// check import route
const (
	importNone importMethod = iota
	importViaKibana
)

// ImportDashboards tries to import the kibana dashboards.
func ImportDashboards(
	homePath string,
	kibanaConfig, dashboardsConfig *common.Config,
) error {
	if dashboardsConfig == nil || !dashboardsConfig.Enabled() {
		return nil
	}

	ctx := context.Background()

	// unpack dashboard config
	dashConfig := defaultConfig
	dashConfig.Beat = "tunipbeat"
	dashConfig.Dir = filepath.Join(homePath, defaultDirectory)
	err := dashboardsConfig.Unpack(&dashConfig)
	if err != nil {
		return err
	}

	// init kibana config object
	if kibanaConfig == nil {
		kibanaConfig = common.NewConfig()
	}

	return setupAndImportDashboardsViaKibana(ctx, "", kibanaConfig, &dashConfig, nil)
}

func setupAndImportDashboardsViaKibana(ctx context.Context, hostname string, kibanaConfig *common.Config,
	dashboardsConfig *Config, msgOutputter MessageOutputter) error {

	kibanaLoader, err := NewKibanaLoader(ctx, kibanaConfig, dashboardsConfig, hostname, msgOutputter)
	if err != nil {
		return fmt.Errorf("fail to create the Kibana loader: %v", err)
	}

	defer kibanaLoader.Close()

	kibanaLoader.statusMsg("Kibana URL %v", kibanaLoader.client.Connection.URL)

	return ImportDashboardsViaKibana(kibanaLoader)
}

func ImportDashboardsViaKibana(kibanaLoader *KibanaLoader) error {
	version, err := common.NewVersion(kibanaLoader.version)
	if err != nil {
		return fmt.Errorf("Invalid Kibana version: %s", kibanaLoader.version)
	}

	importer, err := NewImporter(*version, kibanaLoader.config, kibanaLoader)
	if err != nil {
		return fmt.Errorf("fail to create a Kibana importer for loading the dashboards: %v", err)
	}

	if err := importer.Import(); err != nil {
		return fmt.Errorf("fail to import the dashboards in Kibana: %v", err)
	}

	return nil
}
