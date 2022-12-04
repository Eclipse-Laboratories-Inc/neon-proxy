package configuration

// DATABASES
type StorageConfiguration struct {
	Postgres map[string]*PostgresConfiguration
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
