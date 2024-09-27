package http

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"shortbin/internal/auth/repository"
	"shortbin/internal/auth/service"
	"shortbin/pkg/middleware"
	"shortbin/pkg/validation"
)

func Routes(r *gin.RouterGroup, dbPool *pgxpool.Pool, validator validation.Validation) {
	userRepo := repository.NewUserRepository(dbPool)
	userSvc := service.NewUserService(validator, userRepo)
	userHandler := NewUserHandler(userSvc)

	authMiddleware := middleware.JWTAuth()
	refreshAuthMiddleware := middleware.JWTRefresh()
	authRoute := r.Group("/auth")
	{
		authRoute.POST("/register", userHandler.Register)
		authRoute.POST("/login", userHandler.Login)
		authRoute.POST("/forgot-password", userHandler.ForgotPassword)
		authRoute.POST("/reset-password", userHandler.ResetPassword)
		authRoute.POST("/refresh", refreshAuthMiddleware, userHandler.RefreshToken)
		authRoute.GET("/me", authMiddleware, userHandler.GetMe)
		authRoute.POST("/change-password", authMiddleware, userHandler.ChangePassword)
	}
}
