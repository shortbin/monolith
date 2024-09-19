package http

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"shortbin/internal/retrieve/repository"
	"shortbin/internal/retrieve/service"
	"shortbin/pkg/kafka"
)

func Routes(e *gin.Engine, dbPool *pgxpool.Pool, kafkaProducer kafka.IKafkaProducer) {
	retrieveRepo := repository.NewRetrieveRepository(dbPool)
	retrieveSvc := service.NewRetrieveService(retrieveRepo)
	retrieveHandler := NewRetrieveHandler(retrieveSvc, kafkaProducer)

	e.GET("/:short_id", retrieveHandler.Retrieve)
}
