package config

import (
	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
	"log"
	"path/filepath"
	"runtime"
)

const ProductionEnv = "production"

type Config struct {
	Environment    string `env:"environment"`
	HttpPort       int    `env:"http_port"`
	AuthSecret     string `env:"auth_secret"`
	DataSourceName string `env:"data_source_name"`
}

var (
	cfg Config
)

func LoadConfig() *Config {
	_, filename, _, _ := runtime.Caller(0)
	currentDir := filepath.Dir(filename)

	err := godotenv.Load(filepath.Join(currentDir, "config.yaml"))
	if err != nil {
		log.Printf("Error on load configuration file, error: %v", err)
	}

	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Error on parsing configuration file, error: %v", err)
	}

	return &cfg
}

func GetConfig() *Config {
	return &cfg
}
