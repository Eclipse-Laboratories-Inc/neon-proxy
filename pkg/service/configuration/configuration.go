package configuration

import (
	"fmt"
	"os"
	"strings"

	"github.com/davecgh/go-spew/spew"
)

// SERVICE CONFIGURATION
type ServiceConfiguration struct {
	Name         string
	IsConsoleApp bool
	Storage      *StorageConfiguration
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

	if err = serviceConfiguration.loadStorageConfigurations(cfg.Storage); err != nil {
		return nil, err
	}

	return serviceConfiguration, nil
}

// DATABASES
type StorageConfiguration struct {
	Postgres map[string]*PostgresConfiguration
}

// POSTGRESQL DATABASE
type PostgresConfiguration struct {
	Hostname string
	Port     string
	SSLMode  string
	Username string
	Password string
	Database string
}

func (c *ServiceConfiguration) loadStorageConfigurations(storageList *ConfigStorageList) (err error) {
	if storageList == nil {
		return nil
	}

	if err = c.loadPostgresStorageConfigs(storageList.Postgres); err != nil {
		return err
	}

	return nil
}

func (c *ServiceConfiguration) loadPostgresStorageConfigs(list []string) (err error) {
	for _, db := range list {
		postgresConfig, err := c.loadPostgresStorageConfig(db)
		if err != nil {
			return err
		}
		c.Storage.Postgres[db] = postgresConfig
	}

	return nil
}

// LOAD POSTGRES CONFIGURATION
func (c *ServiceConfiguration) loadPostgresStorageConfig(name string) (cfg *PostgresConfiguration, err error) {
	postgresConfig := &PostgresConfiguration{}

	name = strings.ToUpper(name)

	postgresConfig.Hostname = os.Getenv(fmt.Sprintf("NS_DB_%s_HOSTNAME", name))
	postgresConfig.Port = os.Getenv(fmt.Sprintf("NS_DB_%s_PORT", name))
	postgresConfig.SSLMode = os.Getenv(fmt.Sprintf("NS_DB_%s_SSLMODE", name))
	postgresConfig.Username = os.Getenv(fmt.Sprintf("NS_DB_%s_USERNAME", name))
	postgresConfig.Password = os.Getenv(fmt.Sprintf("NS_DB_%s_PASSWORD", name))
	postgresConfig.Database = os.Getenv(fmt.Sprintf("NS_DB_%s_DATABASE", name))

	if postgresConfig.Port == "" {
		postgresConfig.Port = "5432"
	}

	if postgresConfig.Database == "" || postgresConfig.Hostname == "" || postgresConfig.Username == "" {
		return nil, fmt.Errorf("invalid env parameters for database '%s': %s", name, spew.Sdump(postgresConfig))
	}

	return postgresConfig, nil
}
