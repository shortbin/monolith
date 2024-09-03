package service

import (
	"context"
	"errors"
	"shortbin/internal/auth/model"
	"shortbin/pkg/jwt"
	"shortbin/pkg/utils"

	"golang.org/x/crypto/bcrypt"
	dto "shortbin/internal/auth/dto"
	repository "shortbin/internal/auth/repository"
	"shortbin/pkg/logger"
	"shortbin/pkg/validation"
)

//go:generate mockery --name=IUserService
type IUserService interface {
	Login(ctx context.Context, req *dto.LoginReq) (*model.User, string, string, error)
	Register(ctx context.Context, req *dto.RegisterReq) (*model.User, error)
	GetUserByID(ctx context.Context, id string) (*model.User, error)
	RefreshToken(ctx context.Context, userID string) (string, string, error)
	ChangePassword(ctx context.Context, id string, req *dto.ChangePasswordReq) error
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

func (s *UserService) Login(ctx context.Context, req *dto.LoginReq) (*model.User, string, string, error) {
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, "", "", err
	}

	userDetails, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		logger.Errorf("Login.GetUserByEmail fail, email: %s, error: %s", req.Email, err)
		return nil, "", "", err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(userDetails.Password), []byte(req.Password)); err != nil {
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

func (s *UserService) Register(ctx context.Context, req *dto.RegisterReq) (*model.User, error) {
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, err
	}

	var user model.User
	utils.Copy(&user, &req)
	// populate auth model with values
	user.PopulateValues()
	user.Password = utils.HashAndSalt([]byte(user.Password))
	err := s.repo.Create(ctx, &user)
	if err != nil {
		logger.Errorf("Register.Create fail, email: %s, error: %s", req.Email, err)
		return nil, err
	}
	return &user, nil
}

func (s *UserService) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		logger.Errorf("GetUserByID fail, id: %s, error: %s", id, err)
		return nil, err
	}

	return user, nil
}

func (s *UserService) RefreshToken(ctx context.Context, userID string) (string, string, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		logger.Errorf("RefreshToken.GetUserByID fail, id: %s, error: %s", userID, err)
		return "", "", err
	}

	tokenData := map[string]interface{}{
		"id":    user.ID,
		"email": user.Email,
	}
	accessToken := jwt.GenerateAccessToken(tokenData)
	refreshToken := jwt.GenerateRefreshToken(tokenData)
	return accessToken, refreshToken, nil
}

func (s *UserService) ChangePassword(ctx context.Context, id string, req *dto.ChangePasswordReq) error {
	if err := s.validator.ValidateStruct(req); err != nil {
		return err
	}
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		logger.Errorf("ChangePassword.GetUserByID fail, id: %s, error: %s", id, err)
		return err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return errors.New("wrong password")
	}

	user.Password = utils.HashAndSalt([]byte(req.NewPassword))
	err = s.repo.Update(ctx, user)
	if err != nil {
		logger.Errorf("ChangePassword.Update fail, id: %s, error: %s", id, err)
		return err
	}

	return nil
}
