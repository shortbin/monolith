package config

import (
	"github.com/spf13/viper"
	"log"
)

const ProductionEnv = "production"

type Config struct {
	Environment       string       `mapstructure:"environment"`
	HttpPort          int          `mapstructure:"http_port"`
	AuthSecret        string       `mapstructure:"auth_secret"`
	DataSourceName    string       `mapstructure:"data_source_name"`
	ShortIdLength     ShortIdLimit `mapstructure:"short_id_length"`
	ExpirationInYears int          `mapstructure:"expiration_in_years"`
	Kafka             Kafka        `mapstructure:"kafka"`
	EnablePprof       bool         `mapstructure:"enable_pprof"`
}

type ShortIdLimit struct {
	Default int `mapstructure:"default"`
	Min     int `mapstructure:"min"`
	Max     int `mapstructure:"max"`
}

type Kafka struct {
	Broker string `mapstructure:"broker"`
	Topic  string `mapstructure:"topic"`
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
