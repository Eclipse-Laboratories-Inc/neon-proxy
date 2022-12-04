package configuration

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type MetricsServerConfiguration struct {
	ServiceName   string
	ListenAddress string
	ListenPort    int
	Interval      int
}

// LOAD METRICS CONFIGURATION
func (c *ServiceConfiguration) loadMetricsServerConfiguration(serviceName string) error {
	serviceName = strings.ToUpper(serviceName)
	listenAddr := os.Getenv(fmt.Sprintf("SERVICE_%s_METRICS_LISTEN_ADDRESS", serviceName))
	listenPortString := os.Getenv(fmt.Sprintf("SERVICE_%s_METRICS_LISTEN_PORT", serviceName))
	intervalString := os.Getenv(fmt.Sprintf("SERVICE_%s_METRICS_INTERVAL", serviceName))

	port, err := strconv.Atoi(listenPortString)
	if err != nil {
		port = 0
	}

	interval, err := strconv.Atoi(intervalString)
	if err != nil {
		interval = 0
	}

	cfg := &MetricsServerConfiguration{
		ServiceName:   serviceName,
		ListenAddress: listenAddr,
		ListenPort:    port,
		Interval:      interval,
	}

	c.MetricsServer = cfg

	return nil
}
