package configuration

import (
	"os"
	"strings"
)

type LoggerConfiguration struct {
	Level    string
	UseFile  bool
	FilePath string
}

// LOAD LOGGER CONFIGURATION
func (c *ServiceConfiguration) loadLoggerConfiguration() error {
	var level = os.Getenv("NS_LOG_LEVEL")
	var path = os.Getenv("NS_LOG_PATH")

	if path == "" {
		path = "logs"
	}

	var useFile bool
	var useFileString = strings.ToLower(os.Getenv("NS_LOG_USE_FILE"))
	if useFileString != "" && (useFileString == "true" || useFileString == "t") {
		useFile = true
	}

	cfg := &LoggerConfiguration{
		Level:    level,
		FilePath: path,
		UseFile:  useFile,
	}

	c.Logger = cfg

	return nil
}
