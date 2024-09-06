package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"shortbin/internal/retrieve/service"
	"shortbin/pkg/response"
)

type RetrieveHandler struct {
	service service.IRetrieveService
}

func NewRetrieveHandler(service service.IRetrieveService) *RetrieveHandler {
	return &RetrieveHandler{
		service: service,
	}
}

func (h *RetrieveHandler) Retrieve(c *gin.Context) {
	shortId := c.Param("short_id")

	longUrl, err := h.service.GetLongUrlByShortId(c, shortId)
	if err != nil {
		response.Error(c, http.StatusNotFound, err, "not found")
		return
	}

	c.Redirect(http.StatusMovedPermanently, longUrl)
}
