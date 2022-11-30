package service

import (
	"context"

	"github.com/neonlabsorg/neon-proxy/pkg/logger"
	"github.com/neonlabsorg/neon-proxy/pkg/service/configuration"
)

type DatabaseManager struct {
	ctx                  context.Context
	storageConfiguration *configuration.StorageConfiguration
	logger               logger.Logger
	postgresManager      *PostgresManager
}

func NewDatabaseManager(
	ctx context.Context,
	storageConfiguration *configuration.StorageConfiguration,
	log logger.Logger,
) (manager *DatabaseManager, err error) {
	manager = &DatabaseManager{
		ctx:                  ctx,
		storageConfiguration: storageConfiguration,
		logger:               log,
	}

	manager.postgresManager, err = NewPostgresManager(ctx, log, storageConfiguration.Postgres)
	if err != nil {
		return
	}

	return manager, nil
}

func (m *DatabaseManager) GetPostgresManager() *PostgresManager {
	return m.postgresManager
}
