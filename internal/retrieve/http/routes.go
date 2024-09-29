package http

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"shortbin/internal/retrieve/repository"
	"shortbin/internal/retrieve/service"
	"shortbin/pkg/kafka"
	"shortbin/pkg/redis"
)

func Routes(e *gin.Engine, dbPool *pgxpool.Pool, kafkaProducer kafka.IKafkaProducer, cache redis.IRedis) {
	retrieveRepo := repository.NewRetrieveRepository(dbPool)
	retrieveSvc := service.NewRetrieveService(retrieveRepo)
	retrieveHandler := NewRetrieveHandler(retrieveSvc, kafkaProducer, cache)

	e.GET("/:short_id", retrieveHandler.Retrieve)
}
