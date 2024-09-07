package http

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"shortbin/internal/create/repository"
	"shortbin/internal/create/service"
	"shortbin/pkg/middleware"
	"shortbin/pkg/validation"
)

func Routes(r *gin.RouterGroup, dbPool *pgxpool.Pool, validator validation.Validation) {
	createRepo := repository.NewCreateRepository(dbPool)
	createSvc := service.NewCreateService(validator, createRepo)
	userHandler := NewUserHandler(createSvc)

	authMiddleware := middleware.OptionalJWTAuth()
	r.POST("/create", authMiddleware, userHandler.Create)
}
