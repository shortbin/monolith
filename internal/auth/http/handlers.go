package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"

	"shortbin/internal/auth/dto"
	"shortbin/internal/auth/service"
	"shortbin/pkg/logger"
	"shortbin/pkg/response"
	"shortbin/pkg/utils"
)

type UserHandler struct {
	service service.IUserService
}

func NewUserHandler(service service.IUserService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

// Login godoc
//
//	@Summary	Login
//	@Tags		users
//	@Produce	json
//	@Param		_	body		dto.LoginReq	true	"Body"
//	@Success	200	{object}	dto.LoginRes
//	@Router		/api/v1/auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginReq
	if err := c.ShouldBindJSON(&req); c.Request.Body == nil || err != nil {
		logger.Error("Failed to get body ", err)
		response.Error(c, http.StatusBadRequest, err, response.InvalidParameters)
		return
	}

	user, accessToken, refreshToken, err := h.service.Login(c, &req)
	if err != nil {
		logger.Error("Failed to login ", err)
		response.Error(c, http.StatusBadRequest, err, response.WrongCredentials)
		return
	}

	var res dto.LoginRes
	utils.Copy(&res.User, &user)
	res.AccessToken = accessToken
	res.RefreshToken = refreshToken
	response.JSON(c, http.StatusOK, res)
}

// Register godoc
//
//	@Summary	Register new auth
//	@Tags		users
//	@Produce	json
//	@Param		_	body		dto.RegisterReq	true	"Body"
//	@Success	200	{object}	dto.RegisterRes
//	@Router		/api/v1/auth/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req dto.RegisterReq
	if err := c.ShouldBindJSON(&req); c.Request.Body == nil || err != nil {
		logger.Error("Failed to get body ", err)
		response.Error(c, http.StatusBadRequest, err, response.InvalidParameters)
		return
	}

	user, err := h.service.Register(c, &req)
	if err != nil {
		var pgErr *pgconn.PgError
		// check if error is that user already exists
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			response.Error(c, http.StatusConflict, pgErr, response.UserAlreadyExists)
		} else {
			logger.Error(err.Error())
			response.Error(c, http.StatusInternalServerError, err, response.SomethingWentWrong)
		}
		return
	}

	var res dto.RegisterRes
	utils.Copy(&res, &user)
	response.JSON(c, http.StatusOK, res)
}

// GetMe godoc
//
//	@Summary	get my profile
//	@Tags		users
//	@Security	ApiKeyAuth
//	@Produce	json
//	@Success	200	{object}	dto.User
//	@Router		/api/v1/auth/me [get]
func (h *UserHandler) GetMe(c *gin.Context) {
	userID := c.GetString("userId")
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, errors.New(response.EmptyUserId), response.Unauthorized)
		return
	}

	user, err := h.service.GetUserByID(c, userID)
	if err != nil {
		logger.Error(err.Error())
		response.Error(c, http.StatusInternalServerError, err, response.SomethingWentWrong)
		return
	}

	var res dto.User
	utils.Copy(&res, &user)
	response.JSON(c, http.StatusOK, res)
}

func (h *UserHandler) RefreshToken(c *gin.Context) {
	userID := c.GetString("userId")
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, errors.New(response.EmptyUserId), response.Unauthorized)
		return
	}

	accessToken, err := h.service.RefreshToken(c, userID)
	if err != nil {
		logger.Error("Failed to refresh token ", err)
		response.Error(c, http.StatusUnauthorized, err, response.Unauthorized)
		return
	}

	res := dto.RefreshTokenRes{AccessToken: accessToken}
	response.JSON(c, http.StatusOK, res)
}

// ChangePassword godoc
//
//	@Summary	changes the password
//	@Tags		users
//	@Security	ApiKeyAuth
//	@Produce	json
//	@Param		_	body	dto.ChangePasswordReq	true	"Body"
//	@Router		/api/v1/auth/change-password [put]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	var req dto.ChangePasswordReq
	if err := c.ShouldBindJSON(&req); c.Request.Body == nil || err != nil {
		logger.Error("Failed to get body ", err)
		response.Error(c, http.StatusBadRequest, err, response.InvalidParameters)
		return
	}

	userID := c.GetString("userId")
	err := h.service.ChangePassword(c, userID, &req)
	if err != nil {
		logger.Error(err.Error())
		response.Error(c, http.StatusInternalServerError, err, response.SomethingWentWrong)
		return
	}

	res := map[string]string{"message": "password changed successfully"}
	response.JSON(c, http.StatusOK, res)
}
