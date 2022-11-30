package service

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/gagliardetto/solana-go/rpc"
	"github.com/neonlabsorg/neon-proxy/pkg/logger"
	"github.com/neonlabsorg/neon-proxy/pkg/service/configuration"
	"github.com/urfave/cli/v2"
)

var Version string

type Service struct {
	env             string
	name            string
	version         string
	ctx             context.Context
	cliApp          *cli.App
	cliContext      *cli.Context
	loggerManager   *LoggerManager
	solanaRpcClient *rpc.Client
	handlers        []func(service *Service)
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
		env:     env,
		name:    configuration.Name,
		version: Version,
	}

	s.initContext()
	s.initCliApp(configuration.IsConsoleApp)
	s.initLoggerManager()
	s.initSolana()

	return s
}

func (s *Service) Run() {
	err := s.cliApp.Run(os.Args)
	if err != nil {
		panic(err.Error())
	}
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

func (s *Service) initLoggerManager() {
	var level = os.Getenv("NEON_SERVICE_LOG_LEVEL")
	var path = os.Getenv("NEON_SERVICE_LOG_PATH")

	if level == "" {
		if s.env == "development" {
			level = "debug"
		} else {
			level = "info"
		}
	}

	if path == "" {
		path = "logs"
	}

	var useFile = strings.ToLower(os.Getenv("NEON_SERVICE_LOG_USE_FILE"))

	var log logger.Logger
	var err error
	if useFile != "" && (useFile == "true" || useFile == "t") {
		log, err = logger.NewLogger(s.name, logger.LogSettings{
			Level: strings.ToLower(level),
			Path:  strings.ToLower(path),
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
