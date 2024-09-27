package service

import (
	"errors"

	"github.com/gin-gonic/gin"
	"go.elastic.co/apm/module/apmzap/v2"
	"go.elastic.co/apm/v2"
	"golang.org/x/crypto/bcrypt"

	"shortbin/internal/auth/dto"
	"shortbin/internal/auth/model"
	"shortbin/internal/auth/repository"
	"shortbin/pkg/jwt"
	"shortbin/pkg/logger"
	"shortbin/pkg/utils"
	"shortbin/pkg/validation"
)

//go:generate mockery --name=IUserService
type IUserService interface {
	Login(ctx *gin.Context, req *dto.LoginReq) (*model.User, string, string, error)
	Register(ctx *gin.Context, req *dto.RegisterReq) (*model.User, error)
	GetUserByID(ctx *gin.Context, id string) (*model.User, error)
	RefreshToken(ctx *gin.Context, userID string) (string, error)
	ChangePassword(ctx *gin.Context, userID string, req *dto.ChangePasswordReq) error
	SendPasswordResetEmail(ctx *gin.Context, req *dto.ForgotPasswordReq) (string, error)
	ResetPassword(c *gin.Context, userID string, d *dto.ResetPasswordReq) error
}

type UserService struct {
	validator validation.Validation
	repo      repository.IUserRepository
}

func NewUserService(
	validator validation.Validation,
	repo repository.IUserRepository) *UserService {
	return &UserService{
		validator: validator,
		repo:      repo,
	}
}

func (s *UserService) Login(ctx *gin.Context, req *dto.LoginReq) (*model.User, string, string, error) {
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, "", "", err
	}

	apmTx := apm.TransactionFromContext(ctx.Request.Context())
	traceContextFields := apmzap.TraceContext(ctx.Request.Context())
	rootSpan := apmTx.StartSpan("*UserService.Login", "service", nil)
	defer rootSpan.End()

	userDetails, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		logger.Infof("Login.GetUserByEmail fail, email: %s, error: %s", req.Email, err)
		logger.ApmLogger.With(traceContextFields...).Error(err.Error())
		return nil, "", "", err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(userDetails.HashedPassword), []byte(req.Password)); err != nil {
		return nil, "", "", errors.New("wrong password")
	}

	tokenData := map[string]interface{}{
		"id":    userDetails.ID,
		"email": userDetails.Email,
	}
	accessToken := jwt.GenerateAccessToken(tokenData, jwt.LoginTokenType)
	refreshToken := jwt.GenerateRefreshToken(tokenData)
	return userDetails, accessToken, refreshToken, nil
}

func (s *UserService) Register(ctx *gin.Context, req *dto.RegisterReq) (*model.User, error) {
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, err
	}

	apmTx := apm.TransactionFromContext(ctx.Request.Context())
	traceContextFields := apmzap.TraceContext(ctx.Request.Context())
	rootSpan := apmTx.StartSpan("*UserService.Register", "service", nil)
	defer rootSpan.End()

	var user model.User
	utils.Copy(&user, &req)
	// populate auth model with values
	user.PopulateValues()
	user.HashedPassword = utils.HashAndSalt([]byte(user.HashedPassword))
	err := s.repo.Create(ctx, &user)
	if err != nil {
		logger.Infof("Register.Create fail, email: %s, error: %s", req.Email, err)
		logger.ApmLogger.With(traceContextFields...).Error(err.Error())
		return nil, err
	}
	return &user, nil
}

func (s *UserService) GetUserByID(ctx *gin.Context, id string) (*model.User, error) {
	apmTx := apm.TransactionFromContext(ctx.Request.Context())
	traceContextFields := apmzap.TraceContext(ctx.Request.Context())
	rootSpan := apmTx.StartSpan("*UserService.GetUserByID", "service", nil)
	defer rootSpan.End()

	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		logger.Infof("GetUserByID fail, id: %s, error: %s", id, err)
		logger.ApmLogger.With(traceContextFields...).Error(err.Error())
		return nil, err
	}

	return user, nil
}

func (s *UserService) RefreshToken(ctx *gin.Context, userID string) (string, error) {
	apmTx := apm.TransactionFromContext(ctx.Request.Context())
	traceContextFields := apmzap.TraceContext(ctx.Request.Context())
	rootSpan := apmTx.StartSpan("*UserService.RefreshToken", "service", nil)
	defer rootSpan.End()

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		logger.Infof("RefreshToken.GetUserByID fail, id: %s, error: %s", userID, err)
		logger.ApmLogger.With(traceContextFields...).Error(err.Error())
		return "", err
	}

	tokenData := map[string]interface{}{
		"id":    user.ID,
		"email": user.Email,
	}
	accessToken := jwt.GenerateAccessToken(tokenData, jwt.LoginTokenType)
	return accessToken, nil
}

func (s *UserService) ChangePassword(ctx *gin.Context, userID string, req *dto.ChangePasswordReq) error {
	if err := s.validator.ValidateStruct(req); err != nil {
		return err
	}

	apmTx := apm.TransactionFromContext(ctx.Request.Context())
	traceContextFields := apmzap.TraceContext(ctx.Request.Context())
	rootSpan := apmTx.StartSpan("*UserService.ChangePassword", "service", nil)
	defer rootSpan.End()

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		logger.Infof("ChangePassword.GetUserByID fail, userID: %s, error: %s", userID, err)
		logger.ApmLogger.With(traceContextFields...).Error(err.Error())
		return err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(req.Password)); err != nil {
		return errors.New("wrong password")
	}

	if req.Password == req.NewPassword {
		return errors.New("new password cannot be the same as the old password")
	}

	user.HashedPassword = utils.HashAndSalt([]byte(req.NewPassword))

	err = s.repo.Update(ctx, user)
	if err != nil {
		logger.Infof("ChangePassword.Update fail, userID: %s, error: %s", userID, err)
		logger.ApmLogger.With(traceContextFields...).Error(err.Error())
		return err
	}

	return nil
}

func (s *UserService) SendPasswordResetEmail(ctx *gin.Context, req *dto.ForgotPasswordReq) (string, error) {
	if err := s.validator.ValidateStruct(req); err != nil {
		return "", err
	}

	apmTx := apm.TransactionFromContext(ctx.Request.Context())
	traceContextFields := apmzap.TraceContext(ctx.Request.Context())
	rootSpan := apmTx.StartSpan("*UserService.SendPasswordResetEmail", "service", nil)
	defer rootSpan.End()

	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		logger.Infof("SendPasswordResetEmail.GetUserByEmail fail, email: %s, error: %s", req.Email, err)
		logger.ApmLogger.With(traceContextFields...).Error(err.Error())
		return "", err
	}

	tokenData := map[string]interface{}{
		"id":    user.ID,
		"email": user.Email,
	}
	accessToken := jwt.GenerateAccessToken(tokenData, jwt.ForgotPasswordTokenType)
	return accessToken, nil
}

func (s *UserService) ResetPassword(c *gin.Context, userID string, req *dto.ResetPasswordReq) error {
	if err := s.validator.ValidateStruct(req); err != nil {
		return err
	}

	apmTx := apm.TransactionFromContext(c.Request.Context())
	traceContextFields := apmzap.TraceContext(c.Request.Context())
	rootSpan := apmTx.StartSpan("*UserService.ResetPassword", "service", nil)
	defer rootSpan.End()

	user, err := s.repo.GetUserByID(c, userID)
	if err != nil {
		logger.Infof("ResetPassword.GetUserByID fail, userID: %s, error: %s", userID, err)
		logger.ApmLogger.With(traceContextFields...).Error(err.Error())
		return err
	}

	user.HashedPassword = utils.HashAndSalt([]byte(req.Password))
	err = s.repo.Update(c, user)
	if err != nil {
		logger.Infof("ResetPassword.Update fail, userID: %s, error: %s", userID, err)
		logger.ApmLogger.With(traceContextFields...).Error(err.Error())
		return err
	}

	return nil
}
