package config

import (
	"github.com/spf13/viper"
	"log"
)

const ProductionEnv = "production"

type Config struct {
	Environment    string       `mapstructure:"environment"`
	HttpPort       int          `mapstructure:"http_port"`
	AuthSecret     string       `mapstructure:"auth_secret"`
	DataSourceName string       `mapstructure:"data_source_name"`
	ShortIdLength  ShortIdLimit `mapstructure:"short_id_length"`
}

type ShortIdLimit struct {
	Min int `mapstructure:"min"`
	Max int `mapstructure:"max"`
}

var cfg Config

func LoadConfig(configPath string) *Config {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal("Error reading config file, ", err)
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatal("Unable to decode into struct, ", err)
	}

	return &cfg
}

func GetConfig() *Config {
	return &cfg
}
