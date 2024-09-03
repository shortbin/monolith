package http

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"net/http"

	"shortbin/pkg/config"
	"shortbin/pkg/logger"
	"shortbin/pkg/validation"

	authHttp "shortbin/internal/auth/http"
)

type Server struct {
	engine    *gin.Engine
	cfg       *config.Config
	validator validation.Validation
	db        *pgxpool.Pool
}

func NewServer(validator validation.Validation, db *pgxpool.Pool) *Server {
	return &Server{
		engine:    gin.Default(),
		cfg:       config.GetConfig(),
		validator: validator,
		db:        db,
	}
}

func (s Server) Run() error {
	_ = s.engine.SetTrustedProxies(nil)
	if s.cfg.Environment == config.ProductionEnv {
		gin.SetMode(gin.ReleaseMode)
	}

	if err := s.MapRoutes(); err != nil {
		log.Fatalf("MapRoutes Error: %v", err)
	}

	s.engine.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "online"})
		return
	})

	// APM Middleware
	// s.engine.Use(apmgin.Middleware(s.engine))

	// Start http server
	logger.Info("HTTP server is listening on PORT: ", s.cfg.HttpPort)
	if err := s.engine.Run(fmt.Sprintf(":%d", s.cfg.HttpPort)); err != nil {
		log.Fatalf("Running HTTP server: %v", err)
	}

	return nil
}

func (s Server) GetEngine() *gin.Engine {
	return s.engine
}

func (s Server) MapRoutes() error {
	v1 := s.engine.Group("/api/v1")
	authHttp.Routes(v1, s.db, s.validator)
	return nil
}
