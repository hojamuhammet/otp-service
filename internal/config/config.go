package config

import (
	"log"

	"otp-service/pkg/lib/utils"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string `yaml:"env"`
	Database   `yaml:"database"`
	HTTPServer `yaml:"http_server"`
}

type Database struct {
	Address  string `yaml:"address"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type HTTPServer struct {
	Address string `yaml:"address"`
}

func LoadConfig() *Config {
	configPath := "config.yaml"

	if configPath == "" {
		log.Fatalf("config path is not set or config file does not exist")
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("Cannot read config: %v", utils.Err(err))
	}

	return &cfg
}
