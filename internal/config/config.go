package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type ServiceConfig struct {
	ServerPort string `mapstructure:"server_port"`
}

type CurrencyURL struct {
	URL string `mapstructure:"url"`
}

type DatabaseConfig struct {
	Host           string `mapstructure:"host"`
	Port           int    `mapstructure:"port"`
	User           string `mapstructure:"user"`
	Password       string `mapstructure:"password"`
	Name           string `mapstructure:"name"`
	MigrationsPath string `mapstructure:"migrations_path"`
}

type AppConfig struct {
	Service  ServiceConfig  `mapstructure:"service"`
	Database DatabaseConfig `mapstructure:"database"`
	URL      CurrencyURL    `mapstructure:"currency_url"`
}

func LoadConfig(path string) (AppConfig, error) {
	var config AppConfig

	viper.SetConfigFile(path)

	if err := viper.ReadInConfig(); err != nil {
		return config, fmt.Errorf("error reading config file: %w", err)
	}

	if err := viper.Unmarshal(&config); err != nil {

		return config, fmt.Errorf("error unmarshal config: %w", err)
	}
	return config, nil
}
