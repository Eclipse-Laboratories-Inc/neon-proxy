package metrics

import (
	"context"
	"net/http"
	"time"

	"github.com/neonlabsorg/neon-proxy/pkg/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsServer struct {
	ctx               context.Context
	serviceName       string
	updateIntervalSec int
	listenAddr        string
}

func NewMetricsServer(
	ctx context.Context,
	serviceName string,
	updateIntervalSec int,
	listenAddr string,
) *MetricsServer {
	return &MetricsServer{
		ctx:               ctx,
		serviceName:       serviceName,
		updateIntervalSec: updateIntervalSec,
		listenAddr:        listenAddr,
	}
}

func (s *MetricsServer) Init() error {
	startTime := time.Now()
	uptime := prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "uptime",
		Help:        s.serviceName + " uptime in seconds.",
		ConstLabels: map[string]string{},
	})

	err := prometheus.Register(uptime)
	if err != nil {
		return err
	}
	go func() {
		tick := time.NewTicker(time.Second * time.Duration(s.updateIntervalSec))
		for {
			select {
			case <-s.ctx.Done():
				tick.Stop()
				return
			case <-tick.C:
				uptime.Set(time.Since(startTime).Seconds())
			}
		}
	}()

	return nil
}

func (s *MetricsServer) RunServer() error {
	http.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{Addr: s.listenAddr}

	go func() {
		<-s.ctx.Done()

		if err := srv.Shutdown(context.Background()); err != nil {
			logger.Error().Err(err).Msg("could not shutdown server")
		}
	}()
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}
