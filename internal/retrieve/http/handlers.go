package http

import (
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
// @Param short_id path string true "ShortId ID"
// @Success 301 {string} string "Redirects to the long URL"
// @Failure 404 {object} response.ErrorResponse "Not Found"
// @Router /{short_id} [get]
func (h *RetrieveHandler) Retrieve(c *gin.Context) {
	shortId := c.Param("short_id")

	longUrl, err := h.service.Retrieve(c, shortId)
	if err != nil {
		if e := err.Error(); e == response.IdNotFound || e == response.IdLengthNotInRange {
			response.Error(c, http.StatusNotFound, err, response.IdNotFound)
		} else {
			response.Error(c, http.StatusInternalServerError, err, response.SomethingWentWrong)
		}
		return
	}

	go produce(h, c, shortId)
	c.Redirect(http.StatusMovedPermanently, longUrl)
}

func produce(h *RetrieveHandler, c *gin.Context, shortId string) {
	value := map[string]string{
		"short_id":     shortId,
		"ip_address":   c.ClientIP(),
		"user_agent":   c.GetHeader("User-Agent"),
		"referer":      c.GetHeader("Referer"),
		"request_uri":  c.Request.RequestURI,
		"request_host": c.Request.Host,
	}

	err := h.kafkaProducer.Produce(c, shortId, value)
	if err != nil {
		logger.Infof("failed to produce message to Kafka: %v", err)
		return
	}

	return
}
