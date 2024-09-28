package service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
	RefreshToken(ctx *gin.Context, userID string, userEmail string) (string, error)
	ChangePassword(ctx *gin.Context, userID string, req *dto.ChangePasswordReq) error
	SendPasswordResetEmail(ctx *gin.Context, req *dto.ForgotPasswordReq) (string, error)
	ResetPassword(c *gin.Context, userID string, password string) error
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

	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		logger.Infof("Login.GetUserByEmail fail, email: %s, error: %s", req.Email, err)
		logger.ApmLogger.With(traceContextFields...).Error(err.Error())
		return nil, "", "", err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(req.Password)); err != nil {
		return nil, "", "", errors.New("wrong password")
	}

	tokenData := map[string]interface{}{
		"id":    user.ID,
		"email": user.Email,
	}
	accessToken := jwt.GenerateAccessToken(tokenData, jwt.LoginTokenType)
	refreshToken := jwt.GenerateRefreshToken(tokenData)
	return user, accessToken, refreshToken, nil
}

func (s *UserService) Register(ctx *gin.Context, req *dto.RegisterReq) (*model.User, error) {
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, err
	}

	apmTx := apm.TransactionFromContext(ctx.Request.Context())
	traceContextFields := apmzap.TraceContext(ctx.Request.Context())
	rootSpan := apmTx.StartSpan("*UserService.Register", "service", nil)
	defer rootSpan.End()

	hashedPassword := utils.HashAndSalt([]byte(req.Password))
	user, err := s.repo.Create(ctx, req.Email, hashedPassword)
	if err != nil {
		// Check if the error indicates that the user already exists
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique violation code
			return nil, err
		}
		logger.Infof("Register.Create fail, email: %s, error: %s", req.Email, err)
		logger.ApmLogger.With(traceContextFields...).Error(err.Error())
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetUserByID(ctx *gin.Context, userID string) (*model.User, error) {
	apmTx := apm.TransactionFromContext(ctx.Request.Context())
	traceContextFields := apmzap.TraceContext(ctx.Request.Context())
	rootSpan := apmTx.StartSpan("*UserService.GetUserByID", "service", nil)
	defer rootSpan.End()

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		logger.Infof("GetUserByID fail, userID: %s, error: %s", userID, err)
		logger.ApmLogger.With(traceContextFields...).Error(err.Error())
		return nil, err
	}

	return user, nil
}

func (s *UserService) RefreshToken(ctx *gin.Context, userID string, userEmail string) (string, error) {
	apmTx := apm.TransactionFromContext(ctx.Request.Context())
	rootSpan := apmTx.StartSpan("*UserService.RefreshToken", "service", nil)
	defer rootSpan.End()

	tokenData := map[string]interface{}{
		"id":    userID,
		"email": userEmail,
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
		if errors.Is(err, pgx.ErrNoRows) { // user not found
			return "", err
		}

		logger.Infof("SendPasswordResetEmail.GetUserByEmail fail, email: %s, error: %s", req.Email, err)
		logger.ApmLogger.With(traceContextFields...).Error(err.Error())
		return "", err
	}

	tokenData := map[string]interface{}{
		"id":    user.ID,
		"email": user.Email,
	}
	accessToken := jwt.GenerateAccessToken(tokenData, jwt.ResetPasswordTokenType)
	return accessToken, nil
}

func (s *UserService) ResetPassword(ctx *gin.Context, userID string, password string) error {
	apmTx := apm.TransactionFromContext(ctx.Request.Context())
	traceContextFields := apmzap.TraceContext(ctx.Request.Context())
	rootSpan := apmTx.StartSpan("*UserService.ResetPassword", "service", nil)
	defer rootSpan.End()

	hashedPassword := utils.HashAndSalt([]byte(password))
	err := s.repo.UpdatePassword(ctx, userID, hashedPassword)
	if err != nil {
		logger.Infof("ResetPassword.Update fail, userID: %s, error: %s", userID, err)
		logger.ApmLogger.With(traceContextFields...).Error(err.Error())
		return err
	}

	return nil
}
