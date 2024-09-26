package http

import (
	"go.elastic.co/apm/module/apmzap/v2"
	"net/http"

	"github.com/gin-gonic/gin"

	"shortbin/internal/retrieve/service"
	"shortbin/pkg/kafka"
	"shortbin/pkg/logger"
	"shortbin/pkg/response"
)

type RetrieveHandler struct {
	service       service.IRetrieveService
	kafkaProducer kafka.IKafkaProducer
}

func NewRetrieveHandler(service service.IRetrieveService, kafkaProducer kafka.IKafkaProducer) *RetrieveHandler {
	return &RetrieveHandler{
		service:       service,
		kafkaProducer: kafkaProducer,
	}
}

// Retrieve godoc
//
// @Summary Retrieve a long URL by its short ID
// @Tags urls
// @Produce json
// @Param short_id path string true "Short ID"
// @Success 301 {string} string "Redirects to the long URL"
// @Failure 404 {object} response.ErrorResponse "Not Found"
// @Router /{short_id} [get]
func (h *RetrieveHandler) Retrieve(c *gin.Context) {
	shortID := c.Param("short_id")
	traceContextFields := apmzap.TraceContext(c.Request.Context())

	longURL, err := h.service.Retrieve(c, shortID)
	if err != nil {
		if e := err.Error(); e == response.IDNotFound || e == response.IDLengthNotInRange {
			response.Error(c, http.StatusNotFound, err, response.IDNotFound)
		} else {
			logger.ApmLogger.With(traceContextFields...).Error(err.Error())
			response.Error(c, http.StatusInternalServerError, err, response.SomethingWentWrong)
		}
		return
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
	if err != nil {
		logger.Infof("failed to produce message to Kafka: %v", err)
		return
	}
}
