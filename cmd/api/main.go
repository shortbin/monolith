package main

import (
	httpServer "shortbin/internal/server/http"
	"shortbin/pkg/config"
	"shortbin/pkg/database"
	"shortbin/pkg/logger"
	"shortbin/pkg/validation"
)

func main() {
	cfg := config.LoadConfig("config.yaml")
	logger.Initialize(cfg.Environment)

	db, err := database.NewDatabase(cfg.DataSourceName)
	if err != nil {
		logger.Fatal("Cannot connect to database ", err)
	}

	validator := validation.New()

	httpSvr := httpServer.NewServer(validator, db) //, cache)
	if err = httpSvr.Run(); err != nil {
		logger.Fatal(err)
	}
}
