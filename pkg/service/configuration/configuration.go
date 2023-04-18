package configuration

import (
	"fmt"
	"os"
	"strings"
)

// SERVICE CONFIGURATION
type ServiceConfiguration struct {
	Name                  string
	IsConsoleApp          bool
	GatherStatistics      bool
	Logger                *LoggerConfiguration
	Storage               *StorageConfiguration
	MetricsServer         *MetricsServerConfiguration
	CommunicationProtocol *CommunicationProtocolConfiguration
}

// INIT CONFIGURATION
func NewServiceConfiguration(cfg *Config) (serviceConfiguration *ServiceConfiguration, err error) {
	serviceConfiguration = &ServiceConfiguration{
		Name:         cfg.Name,
		IsConsoleApp: cfg.IsConsoleApp,
		Storage: &StorageConfiguration{
			Postgres: make(map[string]*PostgresConfiguration),
		},
		CommunicationProtocol: &CommunicationProtocolConfiguration{
			RelativeConfigs: make(map[Role][]ProtocolConfiguration),
		},
	}

	gatherStatistics := strings.ToLower(os.Getenv(fmt.Sprintf("NS_%s_GATHER_STATISTICS", cfg.Name)))
	if gatherStatistics == "true" || gatherStatistics == "t" || gatherStatistics == "1" {
		serviceConfiguration.GatherStatistics = true
	}

	if err = serviceConfiguration.loadLoggerConfiguration(); err != nil {
		return nil, err
	}

	if err = serviceConfiguration.loadStorageConfigurations(cfg.Storage); err != nil {
		return nil, err
	}

	if err = serviceConfiguration.loadMetricsServerConfiguration(cfg.Name); err != nil {
		return nil, err
	}

	if err = serviceConfiguration.loadCommunicationProtocolConfiguration(cfg.Name); err != nil {
		return nil, err
	}

	return serviceConfiguration, nil
}
