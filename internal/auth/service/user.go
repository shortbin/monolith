package service

import (
	"errors"

	"github.com/gin-gonic/gin"
	"go.elastic.co/apm/v2"
	"golang.org/x/crypto/bcrypt"
	"shortbin/internal/auth/model"
	"shortbin/pkg/jwt"
	"shortbin/pkg/utils"

	"shortbin/internal/auth/dto"
	"shortbin/internal/auth/repository"
	"shortbin/pkg/logger"
	"shortbin/pkg/validation"
)

//go:generate mockery --name=IUserService
type IUserService interface {
	Login(ctx *gin.Context, req *dto.LoginReq) (*model.User, string, string, error)
	Register(ctx *gin.Context, req *dto.RegisterReq) (*model.User, error)
	GetUserByID(ctx *gin.Context, id string) (*model.User, error)
	RefreshToken(ctx *gin.Context, userID string) (string, error)
	ChangePassword(ctx *gin.Context, id string, req *dto.ChangePasswordReq) error
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

	apmTx := apm.TransactionFromContext(ctx)
	rootSpan := apmTx.StartSpan("*UserService.Login", "service", nil)
	defer rootSpan.End()

	userDetails, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		logger.Infof("Login.GetUserByEmail fail, email: %s, error: %s", req.Email, err)
		return nil, "", "", err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(userDetails.HashedPassword), []byte(req.Password)); err != nil {
		return nil, "", "", errors.New("wrong password")
	}

	tokenData := map[string]interface{}{
		"id":    userDetails.ID,
		"email": userDetails.Email,
	}
	accessToken := jwt.GenerateAccessToken(tokenData)
	refreshToken := jwt.GenerateRefreshToken(tokenData)
	return userDetails, accessToken, refreshToken, nil
}

func (s *UserService) Register(ctx *gin.Context, req *dto.RegisterReq) (*model.User, error) {
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, err
	}

	apmTx := apm.TransactionFromContext(ctx)
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
		return nil, err
	}
	return &user, nil
}

func (s *UserService) GetUserByID(ctx *gin.Context, id string) (*model.User, error) {
	apmTx := apm.TransactionFromContext(ctx.Request.Context())
	rootSpan := apmTx.StartSpan("*UserService.GetUserByID", "service", nil)
	defer rootSpan.End()

	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		logger.Infof("GetUserByID fail, id: %s, error: %s", id, err)
		return nil, err
	}

	return user, nil
}

func (s *UserService) RefreshToken(ctx *gin.Context, userID string) (string, error) {
	apmTx := apm.TransactionFromContext(ctx.Request.Context())
	rootSpan := apmTx.StartSpan("*UserService.RefreshToken", "service", nil)
	defer rootSpan.End()

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		logger.Infof("RefreshToken.GetUserByID fail, id: %s, error: %s", userID, err)
		return "", err
	}

	tokenData := map[string]interface{}{
		"id":    user.ID,
		"email": user.Email,
	}
	accessToken := jwt.GenerateAccessToken(tokenData)
	return accessToken, nil
}

func (s *UserService) ChangePassword(ctx *gin.Context, id string, req *dto.ChangePasswordReq) error {
	if err := s.validator.ValidateStruct(req); err != nil {
		return err
	}

	apmTx := apm.TransactionFromContext(ctx.Request.Context())
	rootSpan := apmTx.StartSpan("*UserService.ChangePassword", "service", nil)
	defer rootSpan.End()

	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		logger.Infof("ChangePassword.GetUserByID fail, id: %s, error: %s", id, err)
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
		logger.Infof("ChangePassword.Update fail, id: %s, error: %s", id, err)
		return err
	}

	return nil
}
