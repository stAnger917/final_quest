package configuration

import (
	"github.com/caarlos0/env/v6"
	flag "github.com/spf13/pflag"
)

type AppConfig struct {
	RunAddress           string `env:"RUN_ADDRESS" envDefault:":8080"`
	DatabaseURI          string `env:"DATABASE_URI" envDefault:"user=postgres password=postgres dbname=gophermart sslmode=disable"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	LoggingLevel         string `env:"LOGGING_LVL" envDefault:"INFO"`
}

func (config *AppConfig) InitAppConfiguration() error {
	if err := env.Parse(config); err != nil {
		return err
	}

	flag.StringVarP(&config.RunAddress, "a", "a", config.RunAddress, "address and port for running app")
	flag.StringVarP(&config.DatabaseURI, "d", "d", config.DatabaseURI, "Database connection string")
	flag.StringVarP(&config.AccrualSystemAddress, "r", "r", config.AccrualSystemAddress, "accrual system connection address")
	flag.StringVarP(&config.LoggingLevel, "l", "l", config.LoggingLevel, "app logging level")
	flag.Parse()

	return nil

}

func NewConfig() *AppConfig {
	return &AppConfig{}
}
