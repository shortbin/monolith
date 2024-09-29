package main

import (
	httpServer "shortbin/internal/server/http"
	"shortbin/pkg/config"
	"shortbin/pkg/database"
	"shortbin/pkg/kafka"
	"shortbin/pkg/logger"
	"shortbin/pkg/redis"
	"shortbin/pkg/validation"
)

func main() {
	cfg := config.LoadConfig("config.yaml")
	logger.Initialize(cfg.Environment)

	db, err := database.NewDatabase(cfg.DataSourceName)
	if err != nil {
		logger.Fatal("Cannot connect to database ", err)
	}

	kp := kafka.NewKafkaProducer(kafka.Config{
		Broker: cfg.Kafka.Broker,
		Topic:  cfg.Kafka.Topic,
	})

	cache := redis.New(redis.Config{
		Address:  cfg.Redis.Address,
		Password: cfg.Redis.Password,
		Database: cfg.Redis.Database,
	})

	validator := validation.New()

	httpSvr := httpServer.NewServer(validator, db, kp, cache)
	if err = httpSvr.Run(); err != nil {
		logger.Fatal(err)
	}
}
