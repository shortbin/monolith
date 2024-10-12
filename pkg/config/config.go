package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

const ProductionEnv = "production"

type Config struct {
	Environment       string       `mapstructure:"environment"`
	HTTPPort          int          `mapstructure:"http_port"`
	AuthSecret        string       `mapstructure:"auth_secret"`
	DataSourceName    string       `mapstructure:"data_source_name"`
	ShortIDLength     ShortIDLimit `mapstructure:"short_id_length"`
	ExpirationInYears int          `mapstructure:"expiration_in_years"`
	Kafka             Kafka        `mapstructure:"kafka"`
	Redis             Redis        `mapstructure:"redis"`
	EnablePprof       bool         `mapstructure:"enable_pprof"`
}

type ShortIDLimit struct {
	Default int `mapstructure:"default"`
	Min     int `mapstructure:"min"`
	Max     int `mapstructure:"max"`
}

type Kafka struct {
	Broker            string `mapstructure:"broker"`
	ClicksTopic       string `mapstructure:"clicks_topic"`
	PublicClicksTopic string `mapstructure:"public_clicks_topic"`
}

type Redis struct {
	Address  string        `mapstructure:"address"`
	Password string        `mapstructure:"password"`
	Database int           `mapstructure:"database"`
	TTL      time.Duration `mapstructure:"ttl"`
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
