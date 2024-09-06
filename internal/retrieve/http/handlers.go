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
		response.Error(c, http.StatusNotFound, err, "not found")
		return
	}

	c.Redirect(http.StatusMovedPermanently, longUrl)
}
