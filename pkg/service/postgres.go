package service

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"github.com/neonlabsorg/neon-proxy/pkg/logger"
	"github.com/neonlabsorg/neon-proxy/pkg/postgres"
	"github.com/neonlabsorg/neon-proxy/pkg/service/configuration"
	"github.com/pkg/errors"
)

type PostgresManager struct {
	configurations map[string]*configuration.PostgresConfiguration
	connectors     map[string]*postgres.Connector
	poolConnectors map[string]*postgres.PoolConnector
	ctx            context.Context
	log            logger.Logger
	mutex          sync.Mutex
}

func NewPostgresManager(
	ctx context.Context,
	log logger.Logger,
	configurations map[string]*configuration.PostgresConfiguration,
) (manager *PostgresManager, err error) {
	manager = &PostgresManager{
		configurations: configurations,
		connectors:     make(map[string]*postgres.Connector),
		log:            log,
		ctx:            ctx,
	}

	if err = manager.init(); err != nil {
		return nil, err
	}

	return manager, nil
}

func (m *PostgresManager) GetDB(dbName string) (connector *postgres.Connector, err error) {
	if len(m.connectors) == 0 {
		err = errors.New("postgres connectors list is empty")
		return
	}

	if dbName == "" {
		for _, connector = range m.connectors {
			return
		}
	}

	connector, err = m.getConnector(dbName)
	if err != nil {
		return
	}

	return
}

func (m *PostgresManager) GetPool(dbName string) (poolConnector *postgres.PoolConnector, err error) {
	if dbName == "" {
		for _, poolConnector = range m.poolConnectors {
			return
		}
	}

	return m.getPoolConnector(dbName)
}

func (m *PostgresManager) getConnector(name string) (connector *postgres.Connector, err error) {
	connector, ok := m.connectors[name]

	if ok && connector != nil {
		err = m.pingDB(connector)
		if err == nil {
			return
		}
	}

	connector, err = m.createConnector(name)
	if err != nil {
		return
	}

	m.addConnector(name, connector)

	return
}

func (m *PostgresManager) getPoolConnector(name string) (poolConnector *postgres.PoolConnector, err error) {
	poolConnector, ok := m.poolConnectors[name]

	if ok && poolConnector != nil {
		err = m.pingPool(poolConnector)
		if err == nil {
			return
		}
	}

	poolConnector, err = m.createPoolConnector(name)
	if err != nil {
		return
	}

	m.addPoolConnector(name, poolConnector)

	return
}

func (m *PostgresManager) init() error {
	for name := range m.configurations {
		connector, err := m.createConnector(name)
		if err != nil {
			return errors.Wrapf(err, "connecting to postgres %s", name)
		}
		m.addConnector(name, connector)

		poolConnector, err := m.createPoolConnector(name)
		if err != nil {
			return errors.Wrapf(err, "connecting to postgres %s", name)
		}
		m.addPoolConnector(name, poolConnector)
	}

	return nil
}

func (m *PostgresManager) addConnector(name string, connector *postgres.Connector) {
	m.mutex.Lock()
	m.connectors[name] = connector
	m.mutex.Unlock()
}

func (m *PostgresManager) addPoolConnector(name string, poolConnector *postgres.PoolConnector) {
	m.mutex.Lock()
	m.poolConnectors[name] = poolConnector
	m.mutex.Unlock()
}

func (m *PostgresManager) createConnector(name string) (connector *postgres.Connector, err error) {
	cfg, err := m.getConfiguration(name)
	if err != nil {
		return
	}

	connectionString := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s password=%s",
		cfg.Hostname, cfg.Port, cfg.Username, cfg.Database, cfg.SSLMode, cfg.Password)

	connection, err := sql.Open("postgres", connectionString)
	if err != nil {
		return
	}

	connector = postgres.NewPostgresConnector(name, m.ctx, connection, cfg)

	err = m.pingDB(connector)
	if err != nil {
		return
	}

	return
}

func (m *PostgresManager) createPoolConnector(name string) (poolConnector *postgres.PoolConnector, err error) {
	cfg, err := m.getConfiguration(name)
	if err != nil {
		return
	}

	connectionString := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s password=%s",
		cfg.Hostname, cfg.Port, cfg.Username, cfg.Database, cfg.SSLMode, cfg.Password)

	connPool, err := pgxpool.New(m.ctx, connectionString)
	if err != nil {
		return
	}

	poolConnector = postgres.NewPoolConnector(name, m.ctx, connPool, cfg)
	err = m.pingPool(poolConnector)
	return
}

func (m *PostgresManager) getConfiguration(name string) (config *configuration.PostgresConfiguration, err error) {
	config, ok := m.configurations[name]

	if !ok {
		err = fmt.Errorf(fmt.Sprintf("configuration for connection %s not found", name))
		return
	}

	return
}

func (m *PostgresManager) pingDB(connector *postgres.Connector) error {
	err := connector.DB.Ping()

	if err != nil {
		return err
	}

	return nil
}

func (m *PostgresManager) pingPool(poolConnector *postgres.PoolConnector) error {
	err := poolConnector.Pool.Ping(m.ctx)

	if err != nil {
		return err
	}

	return nil
}

func (m *PostgresManager) ShutDown() {
	// close all db connections
	for _, connector := range m.connectors {
		connector.GetRawDB().Close()
	}

	// close all db pool connections
	for _, poolConnector := range m.poolConnectors {
		poolConnector.GetPool().Close()
	}
}
