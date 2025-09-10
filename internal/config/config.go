package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type ServiceConfig struct {
	ServerPort string `mapstructure:"server_port"`
}

type CurrencyInfo struct {
	URL    string `mapstructure:"url"`
	Target string `mapstructure:"target"`
}

type DatabaseConfig struct {
	Host           string `mapstructure:"host"`
	Port           int    `mapstructure:"port"`
	User           string `mapstructure:"user"`
	Password       string `mapstructure:"password"`
	Name           string `mapstructure:"name"`
	MigrationsPath string `mapstructure:"migrations_path"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type KafkaConfig struct {
	BrokerHost     string `mapstructure:"host_broker"`
	ControllerHost string `mapstructure:"host_controller"`
}

type AppConfig struct {
	Service  *ServiceConfig  `mapstructure:"service"`
	Database *DatabaseConfig `mapstructure:"database"`
	Currency *CurrencyInfo   `mapstructure:"currency"`
	Redis    *RedisConfig    `mapstructure:"redis"`
	Kafka    *KafkaConfig    `mapstructure:kafka`
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
