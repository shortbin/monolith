package http

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.elastic.co/apm/module/apmgin/v2"

	authHttp "shortbin/internal/auth/http"
	createHttp "shortbin/internal/create/http"
	retrieveHttp "shortbin/internal/retrieve/http"
	"shortbin/pkg/config"
	"shortbin/pkg/kafka"
	"shortbin/pkg/logger"
	"shortbin/pkg/redis"
	"shortbin/pkg/validation"
)

type Server struct {
	engine    *gin.Engine
	cfg       *config.Config
	validator validation.Validation
	db        *pgxpool.Pool
	kp        kafka.IKafkaProducer
	cache     redis.IRedis
}

func NewServer(
	validator validation.Validation,
	db *pgxpool.Pool,
	kp kafka.IKafkaProducer,
	cache redis.IRedis,
) *Server {
	return &Server{
		engine:    gin.Default(),
		cfg:       config.GetConfig(),
		validator: validator,
		db:        db,
		kp:        kp,
		cache:     cache,
	}
}

func (s Server) Run() error {
	_ = s.engine.SetTrustedProxies(nil)
	if s.cfg.Environment == config.ProductionEnv {
		gin.SetMode(gin.ReleaseMode)
	}

	if s.cfg.EnablePprof {
		pprof.Register(s.engine)
	}

	s.engine.Use(apmgin.Middleware(s.engine))

	if err := s.MapRoutes(); err != nil {
		log.Fatalf("MapRoutes Error: %v", err)
	}

	s.engine.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "online"})
	})

	// Start http server
	logger.Info("HTTP server is listening on port ", s.cfg.HTTPPort)
	if err := s.engine.Run(fmt.Sprintf(":%d", s.cfg.HTTPPort)); err != nil {
		log.Fatalf("Running HTTP server: %v", err)
	}

	return nil
}

func (s Server) GetEngine() *gin.Engine {
	return s.engine
}

func (s Server) MapRoutes() error {
	v1 := s.engine.Group("/api/v1")

	retrieveHttp.Routes(s.engine, s.db, s.kp, s.cache)
	authHttp.Routes(v1, s.db, s.validator)
	createHttp.Routes(v1, s.db, s.validator)

	return nil
}
