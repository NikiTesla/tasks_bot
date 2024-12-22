package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Debug          bool            `envconfig:"DEBUG" default:"false"`
	Local          bool            `envconfig:"LOCAL" default:"false"`
	PostgresConfig *PostgresConfig `envconfig:"POSTGRES"`
	TelegramConfig *TelegramConfig `envconfig:"TELEGRAM"`
}

type TelegramConfig struct {
	Debug                bool   `envconfig:"DEBUG" default:"false"`
	APIToken             string `envconfig:"API_TOKEN" required:"true"`
	AdminID              int64  `envconfig:"ADMIN_ID"`
	AdminUsername        string `envconfig:"ADMIN_USERNAME"`
	ChiefPasswordHash    string `envconfig:"CHIEF_PASSWORD_HASH"`
	ExecutorPasswordHash string `envconfig:"EXECUTOR_PASSWORD_HASH"`
	ObserverPasswordHash string `envconfig:"OBSERVER_PASSWORD_HASH"`
	AdminPasswordHash    string `envconfig:"ADMIN_PASSWORD_HASH"`
}

type PostgresConfig struct {
	DSN string `envconfig:"DSN"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := envconfig.Process("", cfg); err != nil {
		return nil, fmt.Errorf("process load config: %w", err)
	}
	return cfg, nil
}
