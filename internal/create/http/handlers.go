package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"shortbin/internal/create/dto"
	"shortbin/internal/create/service"
	"shortbin/pkg/logger"
	"shortbin/pkg/response"
	"shortbin/pkg/utils"
)

type CreateHandler struct {
	service service.ICreateService
}

func NewUserHandler(service service.ICreateService) *CreateHandler {
	return &CreateHandler{
		service: service,
	}
}

func (h CreateHandler) Create(c *gin.Context) {
	var req dto.CreateReq
	if err := c.ShouldBindJSON(&req); c.Request.Body == nil || err != nil {
		logger.Error("Failed to get body ", err)
		response.Error(c, http.StatusBadRequest, err, response.InvalidParameters)
		return
	}

	userID := c.GetString("userID")
	url, err := h.service.Create(c, userID, &req)
	if err != nil {
		logger.Error(err.Error())
		response.Error(c, http.StatusInternalServerError, err, response.SomethingWentWrong)
		return
	}

	var res dto.CreateRes
	utils.Copy(&res, &url)
	response.JSON(c, http.StatusOK, res)
}
