package service

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/gagliardetto/solana-go/rpc"
	"github.com/neonlabsorg/neon-proxy/pkg/communication"
	"github.com/neonlabsorg/neon-proxy/pkg/event"
	"github.com/neonlabsorg/neon-proxy/pkg/logger"
	"github.com/neonlabsorg/neon-proxy/pkg/metrics"
	"github.com/neonlabsorg/neon-proxy/pkg/postgres"
	"github.com/neonlabsorg/neon-proxy/pkg/service/configuration"
	"github.com/urfave/cli/v2"
)

type PostServiceStartedEvent struct {
	serviceName string
}

func (e PostServiceStartedEvent) Name() string {
	return "post.service.created"
}

func (e PostServiceStartedEvent) IsAsynchronous() bool {
	return false
}

func (e PostServiceStartedEvent) ServiceName() string {
	return e.serviceName
}

type PostServiceOnlineEvent struct {
	serviceName string
}

func (e PostServiceOnlineEvent) Name() string {
	return "post.service.online"
}

func (e PostServiceOnlineEvent) IsAsynchronous() bool {
	return false
}

func (e PostServiceOnlineEvent) ServiceName() string {
	return e.serviceName
}

var Version string

type Service struct {
	env              string
	name             string
	version          string
	gatherStatistics bool
	ctx              context.Context
	cliApp           *cli.App
	cliContext       *cli.Context
	loggerManager    *LoggerManager
	dbManager        *DatabaseManager
	solanaRpcClient  *rpc.Client
	handlers         []func(service *Service)
}

func CreateService(
	cfg *configuration.Config,
) *Service {
	configuration, err := configuration.NewServiceConfiguration(cfg)
	if err != nil {
		panic(err)
	}

	env := os.Getenv("NEON_SERVICE_ENV")
	if env == "" {
		env = "development"
	}

	if Version == "v." {
		Version = "v0.0.1"
	}

	s := &Service{
		env:              env,
		name:             configuration.Name,
		version:          Version,
		gatherStatistics: configuration.GatherStatistics,
	}

	s.initContext()
	s.initCliApp(configuration.IsConsoleApp)
	s.initLoggerManager(configuration.Logger)
	s.initSolana()

	if !configuration.IsConsoleApp {
		s.initMetrics(configuration.MetricsServer)
	}

	s.initCommunicationProtocol(configuration.CommunicationProtocol)

	s.dbManager, err = NewDatabaseManager(s.GetContext(), configuration.Storage, s.GetLogger())
	if err != nil {
		panic(err)
	}

	postCreated := PostServiceStartedEvent{serviceName: s.name}
	event.DispatcherInstance().Notify(postCreated)

	return s
}

func (s *Service) Run() {
	err := s.cliApp.Run(os.Args)
	if err != nil {
		panic(err.Error())
	}

	onlineEvent := PostServiceOnlineEvent{serviceName: s.name}
	event.DispatcherInstance().Notify(onlineEvent)
}

func (s *Service) run(cliContext *cli.Context) (err error) {
	s.cliContext = cliContext
	s.loggerManager.GetLogger().Info().Msgf("Service %s version %s started", s.name, s.version)

	var wg sync.WaitGroup
	wg.Add(len(s.handlers))

	for _, handler := range s.handlers {
		go func(h func(s *Service), wGroup *sync.WaitGroup) {
			defer wGroup.Done()
			h(s)
		}(handler, &wg)
	}

	<-s.ctx.Done()
	wg.Wait()

	s.loggerManager.GetLogger().Info().Msgf("Service %s has been stopped", s.name)

	return
}

func (s *Service) initSolana() {
	solanaURL := os.Getenv("NS_SOLANA_URL")
	s.solanaRpcClient = rpc.New(solanaURL)
}

func (s *Service) initContext() {
	ctx, cancel := context.WithCancel(context.Background())
	sigquit := make(chan os.Signal, 1)
	signal.Ignore(syscall.SIGHUP, syscall.SIGPIPE)
	signal.Notify(sigquit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigquit
		cancel()
	}()

	s.ctx = ctx
}

func (s *Service) initLoggerManager(cfg *configuration.LoggerConfiguration) {
	if cfg.Level == "" {
		if s.env == "development" {
			cfg.Level = "debug"
		} else {
			cfg.Level = "info"
		}
	}

	var log logger.Logger
	var err error

	if cfg.UseFile {
		log, err = logger.NewLogger(s.name, logger.LogSettings{
			Level: strings.ToLower(cfg.Level),
			Path:  strings.ToLower(cfg.FilePath),
		})

		if err != nil {
			panic(err)
		}
	} else {
		logger.InitDefaultLogger()
		log = logger.Get()
	}

	logger.SetDefaultLogger(log)

	s.loggerManager = NewLoggerManager(log)
}

func (s *Service) initCliApp(isConsoleApp bool) {
	s.cliApp = cli.NewApp()
	s.cliApp.Name = s.name
	s.cliApp.Version = s.version

	if !isConsoleApp {
		s.cliApp.Action = s.run
	}
}

func (s *Service) initMetrics(cfg *configuration.MetricsServerConfiguration) {
	if cfg.ListenAddress == "" || cfg.ListenPort == 0 || cfg.Interval == 0 {
		s.GetLogger().Info().Msg("Metrics server inicialization has been skipped")
		return
	}

	metricsServer := metrics.NewMetricsServer(
		s.GetContext(),
		cfg.ServiceName,
		cfg.Interval,
		fmt.Sprintf("%s:%d", cfg.ListenAddress, cfg.ListenPort),
	)

	if err := metricsServer.Init(); err != nil {
		s.GetLogger().Error().Err(err).Msg("can't initialize metrics")
		panic(err)
	}

	go func() {
		if err := metricsServer.RunServer(); err != nil {
			s.GetLogger().Error().Err(err).Msg("can't start metrics server")
			panic(err)
		}
	}()
}

func (s *Service) initCommunicationProtocol(cfg *configuration.CommunicationProtocolConfiguration) {
	if cfg.MainConfig == nil || cfg.CommunicationServerPort == "" || cfg.CommunicationEndpointServerPort == "" {
		s.GetLogger().Info().Msg("Communication Protocol server inicialization has been skipped")
		return
	}

	communicationServer := communication.NewCommunicationServer(
		s.GetContext(),
		cfg.MainConfig.Role,
		cfg.MainConfig.InstanceID,
		cfg.MainConfig.Ip,
		cfg.MainConfig.Cluster,
		cfg.RelativeConfigs,
		cfg.CommunicationServerPort,
		cfg.CommunicationEndpointServerPort,
	)

	go func() {
		if err := communicationServer.RegisterServer(); err != nil {
			s.GetLogger().Error().Err(err).Msg("can't register communication server")
			panic(err)
		}
	}()

	if err := communicationServer.Init(); err != nil {
		s.GetLogger().Error().Err(err).Msg("can't initialize communication protocols")
		panic(err)
	}

	go func() {
		if err := communicationServer.RunServer(); err != nil {
			s.GetLogger().Error().Err(err).Msg("can't start communication server")
			panic(err)
		}
	}()
}

func (s *Service) ShutDown() {
	// todo Shut Down all needed services
	s.dbManager.ShutDown()
}

func (s *Service) ModifyCliApp(handler func(cliApp *cli.App)) {
	handler(s.cliApp)
}

func (s *Service) AddHandler(handler func(service *Service)) {
	s.handlers = append(s.handlers, handler)
}

func (s *Service) GetName() string {
	return s.name
}

func (s *Service) GetEnvironment() string {
	return s.env
}

func (s *Service) GetContext() context.Context {
	return s.ctx
}

func (s *Service) GetLogger() logger.Logger {
	return s.loggerManager.GetLogger()
}

func (s *Service) GetSolanaRpcClient() *rpc.Client {
	return s.solanaRpcClient
}

func (s *Service) GetDB(name string) (*postgres.Connector, error) {
	// here you can get any DB from any DBManager (postgress, clickhouse, mysql...)
	return s.dbManager.GetPostgresManager().GetDB(name)
}

func (s *Service) GetPool(name string) (*postgres.PoolConnector, error) {
	return s.dbManager.GetPostgresManager().getPoolConnector(name)
}

func (s *Service) GatherStatistics() bool {
	return s.gatherStatistics
}
