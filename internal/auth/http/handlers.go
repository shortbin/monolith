package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"shortbin/internal/auth/dto"
	"shortbin/internal/auth/service"
	"shortbin/pkg/jwt"
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
		// Check for user already exists error
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			response.Error(c, http.StatusConflict, err, response.UserAlreadyExists)
			return
		}

		logger.Error(err.Error())
		response.Error(c, http.StatusInternalServerError, err, response.SomethingWentWrong)
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
		response.Error(c, http.StatusUnauthorized, errors.New(response.EmptyUserID), response.Unauthorized)
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

// RefreshToken godoc
//
//	@Summary	changes the password
//	@Tags		users
//	@Security	ApiKeyAuth
//	@Produce	json
//	@Success	200	{object} 			dto.RefreshTokenRes
//	@Router		/api/v1/auth/refresh 	[post]
func (h *UserHandler) RefreshToken(c *gin.Context) {
	userID := c.GetString("userId")
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, errors.New(response.EmptyUserID), response.Unauthorized)
		return
	}

	userEmail := c.GetString("userEmail")
	accessToken, err := h.service.RefreshToken(c, userID, userEmail)
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

// ForgotPassword godoc
//
//	@Summary	forgot password
//	@Tags		users
//	@Produce	json
//	@Param		_	body	dto.ForgotPasswordReq	true	"Body"
//	@Router		/api/v1/auth/forgot-password [post]
func (h *UserHandler) ForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordReq
	if err := c.ShouldBindJSON(&req); c.Request.Body == nil || err != nil {
		logger.Error("Failed to get body ", err)
		response.Error(c, http.StatusBadRequest, err, response.InvalidParameters)
		return
	}

	// accessToken is temporary and should be sent over email
	accessToken, err := h.service.SendPasswordResetEmail(c, &req)
	if err != nil {
		// check if error is that userID not found
		if errors.Is(err, pgx.ErrNoRows) {
			response.Error(c, http.StatusNotFound, err, response.UserNotFound)
		} else {
			logger.Error(err)
			response.Error(c, http.StatusInternalServerError, err, response.SomethingWentWrong)
		}
		return
	}

	res := map[string]string{
		"message":      "forgot password email sent",
		"access_token": accessToken,
	}
	response.JSON(c, http.StatusOK, res)
}

// ResetPassword godoc
//
//	@Summary	reset password
//	@Tags		users
//	@Produce	json
//	@Param		_	body	dto.ResetPasswordReq	true	"Body"
//	@Router		/api/v1/auth/reset-password [post]
func (h *UserHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordReq
	if err := c.ShouldBindJSON(&req); c.Request.Body == nil || err != nil {
		logger.Error("Failed to get body ", err)
		response.Error(c, http.StatusBadRequest, err, response.InvalidParameters)
		return
	}

	payload, err := jwt.ValidateToken(req.ResetToken)
	if err != nil || payload == nil || payload["type"] != jwt.ResetPasswordTokenType {
		c.JSON(http.StatusUnauthorized, nil)
		c.Abort()
		return
	}

	userID := payload["id"].(string)
	err = h.service.ResetPassword(c, userID, req.Password)

	if err != nil {
		logger.Error(err.Error())
		response.Error(c, http.StatusInternalServerError, err, response.SomethingWentWrong)
		return
	}

	res := map[string]string{"message": "password reset successful"}
	response.JSON(c, http.StatusOK, res)
}
