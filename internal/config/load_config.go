package config

import (
	"fmt"
	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigdotenv"
	"os"
	"path"
	"reflect"
)

type Config interface {
	Validate() error
}

type ConfigOption func(*ConfigOptions)

type ConfigOptions struct {
	EnvPath    string
	FileName   string
	Validation bool
}

func WithEnvPath(v string) ConfigOption {
	return func(o *ConfigOptions) {
		o.EnvPath = v
	}
}

func WithFileName(v string) ConfigOption {
	return func(o *ConfigOptions) {
		o.FileName = v
	}
}

func WithValidation(v bool) ConfigOption {
	return func(o *ConfigOptions) {
		o.Validation = v
	}
}

func LoadConfigFromEnv(cfg Config, opts ...ConfigOption) error {
	if reflect.ValueOf(cfg).Kind() != reflect.Ptr {
		return fmt.Errorf("config variable must be a pointer")
	}

	options := ConfigOptions{
		Validation: true,
	}

	for _, opt := range opts {
		opt(&options)
	}

	if options.EnvPath == "" {
		pwdDir, err := os.Getwd()
		if err != nil {
			return err
		}
		options.EnvPath = pwdDir
	}

	fileName := ".env"
	if options.FileName != "" {
		fileName = options.FileName
	}

	aconf := aconfig.Config{
		AllowUnknownFields: true,
		SkipFlags:          true,
		SkipDefaults:       false,
		Files:              []string{path.Join(options.EnvPath, fileName)},
		FileDecoders: map[string]aconfig.FileDecoder{
			".env": aconfigdotenv.New(),
		},
	}

	loader := aconfig.LoaderFor(cfg, aconf)
	if err := loader.Load(); err != nil {
		return err
	}

	if !options.Validation {
		return nil
	}

	return cfg.Validate()
}
