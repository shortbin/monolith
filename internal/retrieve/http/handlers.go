package http

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.elastic.co/apm/module/apmzap/v2"

	"shortbin/internal/retrieve/service"
	"shortbin/pkg/config"
	"shortbin/pkg/kafka"
	"shortbin/pkg/logger"
	"shortbin/pkg/redis"
	"shortbin/pkg/response"
)

type RetrieveHandler struct {
	service       service.IRetrieveService
	kafkaProducer kafka.IKafkaProducer
	redis         redis.IRedis
}

func NewRetrieveHandler(service service.IRetrieveService, kafkaProducer kafka.IKafkaProducer, redis redis.IRedis) *RetrieveHandler {
	return &RetrieveHandler{
		service:       service,
		kafkaProducer: kafkaProducer,
		redis:         redis,
	}
}

// Retrieve godoc
//
// @Summary Retrieve a long URL by its short ID
// @Tags urls
// @Produce json
// @Param short_id path string true "Short ID"
// @Success 301 {string} string "Redirects to the long URL"
// @Failure 404 {object} response.ErrorResponse "id not Found"
// @Router /{short_id} [get]
func (h *RetrieveHandler) Retrieve(c *gin.Context) {
	shortID := c.Param("short_id")
	traceContextFields := apmzap.TraceContext(c.Request.Context())

	var longURL string // empty ""
	if err := h.redis.GetByRefreshingExpiry(shortID, &longURL); !errors.Is(err, redis.NilReturn) && err != nil {
		logger.ApmLogger.With(traceContextFields...).Error(err.Error())
	}

	if longURL == "" {
		var err error
		longURL, err = h.service.Retrieve(c, shortID)
		if err != nil {
			if e := err.Error(); e == response.IDNotFound || e == response.IDLengthNotInRange {
				response.Error(c, http.StatusNotFound, err, response.IDNotFound)
			} else {
				logger.ApmLogger.With(traceContextFields...).Error(err.Error())
				response.Error(c, http.StatusInternalServerError, err, response.SomethingWentWrong)
			}
			return
		}
		go cache(h, c, shortID, longURL)
	}
	go produce(h, c, shortID)
	c.Redirect(http.StatusMovedPermanently, longURL)
}

func produce(h *RetrieveHandler, c *gin.Context, shortID string) {
	value := map[string]string{
		"short_id":        shortID,
		"ip_address":      c.ClientIP(),
		"user_agent":      c.GetHeader("User-Agent"),
		"referer":         c.GetHeader("Referer"),
		"x_forwarded_for": c.GetHeader("X-Forwarded-For"),
		"request_host":    c.Request.Host,
	}

	err := h.kafkaProducer.Produce(c, shortID, value)
	traceContextFields := apmzap.TraceContext(c.Request.Context())
	if err != nil {
		logger.Infof("failed to produce message to Kafka: %v", err)
		logger.ApmLogger.With(traceContextFields...).Error(err.Error())
	}
}

func cache(h *RetrieveHandler, c *gin.Context, shortID, longURL string) {
	traceContextFields := apmzap.TraceContext(c.Request.Context())
	if err := h.redis.Set(shortID, longURL, config.GetConfig().Redis.TTL*time.Minute); err != nil {
		logger.Infof("failed to set cache: %v", err)
		logger.ApmLogger.With(traceContextFields...).Error(err.Error())
	}
}
