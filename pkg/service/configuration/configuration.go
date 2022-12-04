package configuration

// SERVICE CONFIGURATION
type ServiceConfiguration struct {
	Name          string
	IsConsoleApp  bool
	Logger        *LoggerConfiguration
	Storage       *StorageConfiguration
	MetricsServer *MetricsServerConfiguration
}

// INIT CONFIGURATION
func NewServiceConfiguration(cfg *Config) (serviceConfiguration *ServiceConfiguration, err error) {
	serviceConfiguration = &ServiceConfiguration{
		Name:         cfg.Name,
		IsConsoleApp: cfg.IsConsoleApp,
		Storage: &StorageConfiguration{
			Postgres: make(map[string]*PostgresConfiguration),
		},
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

	return serviceConfiguration, nil
}
