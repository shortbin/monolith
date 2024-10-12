package http

import (
	"errors"
	"net/http"
	"strings"
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

	var value string // inits to ""
	if err := h.redis.GetByRefreshingExpiry(shortID, &value); !errors.Is(err, redis.NilReturn) && err != nil {
		logger.ApmLogger.With(traceContextFields...).Error(err.Error())
	}

	var longURL string
	var userID string
	if value == "" {
		var err error
		var subUserID *string
		longURL, subUserID, err = h.service.Retrieve(c, shortID)
		if err != nil {
			if e := err.Error(); e == response.IDNotFound || e == response.IDLengthNotInRange {
				response.Error(c, http.StatusNotFound, err, response.IDNotFound)
			} else {
				logger.ApmLogger.With(traceContextFields...).Error(err.Error())
				response.Error(c, http.StatusInternalServerError, err, response.SomethingWentWrong)
			}
			return
		}
		if userID = "-1"; subUserID != nil {
			userID = *subUserID
		}
		go cache(h, c, shortID, userID, longURL)
	} else {
		split := strings.SplitN(value, ";", 2)
		userID, longURL = split[0], split[1]
	}
	go produce(h, c, shortID, userID, longURL)
	c.Redirect(http.StatusMovedPermanently, longURL)
}

func produce(h *RetrieveHandler, c *gin.Context, shortID string, shortCreatedBy string, longURL string) {
	value := map[string]string{
		"short_id":         shortID,
		"short_created_by": shortCreatedBy,
		"long_url":         longURL,
		"ip_address":       c.ClientIP(),
		"user_agent":       c.GetHeader("User-Agent"),
		"referer":          c.GetHeader("Referer"),
		"x_forwarded_for":  c.GetHeader("X-Forwarded-For"),
		"request_host":     c.Request.Host,
	}

	var err error
	if shortCreatedBy == "-1" {
		err = h.kafkaProducer.Produce(c, config.GetConfig().Kafka.PublicClicksTopic, shortID, value)
	} else {
		err = h.kafkaProducer.Produce(c, config.GetConfig().Kafka.ClicksTopic, shortID, value)
	}
	traceContextFields := apmzap.TraceContext(c.Request.Context())
	if err != nil {
		logger.Infof("failed to produce message to Kafka: %v", err)
		logger.ApmLogger.With(traceContextFields...).Error(err.Error())
	}
}

func cache(h *RetrieveHandler, c *gin.Context, shortID string, userID string, longURL string) {
	traceContextFields := apmzap.TraceContext(c.Request.Context())
	if err := h.redis.Set(shortID, userID+";"+longURL, config.GetConfig().Redis.TTL*time.Minute); err != nil {
		logger.Infof("failed to set cache: %v", err)
		logger.ApmLogger.With(traceContextFields...).Error(err.Error())
	}
}
