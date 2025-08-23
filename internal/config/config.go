package config

import (
	"demoserv/internal/models"
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Postgres   models.PostgresConfig   `yaml:"POSTGRES"`
	Kafka      models.KafkaConfig      `yaml:"KAFKA"`
	HttpServer models.HttpServerConfig `yaml:"HTTP_SERVER"`
}

// New конфиг
func New() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadConfig("config/config.yaml", &cfg); err != nil {
		return nil, fmt.Errorf("config error: %v", err)
	}
	return &cfg, nil
}
