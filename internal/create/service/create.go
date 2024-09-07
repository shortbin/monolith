package service

import (
	"context"
	"shortbin/internal/common/model"
	"shortbin/internal/create/dto"
	"shortbin/internal/create/repository"
	"shortbin/pkg/config"
	"shortbin/pkg/logger"
	"shortbin/pkg/utils"
	"shortbin/pkg/validation"
)

//go:generate mockery --name=ICreateService
type ICreateService interface {
	Create(ctx context.Context, id string, req *dto.CreateReq) (*model.Url, error)
}

type CreateService struct {
	validator validation.Validation
	repo      repository.ICreateRepository
}

func NewCreateService(
	validator validation.Validation,
	repo repository.ICreateRepository) *CreateService {
	return &CreateService{
		validator: validator,
		repo:      repo,
	}
}

func (s CreateService) Create(ctx context.Context, id string, req *dto.CreateReq) (*model.Url, error) {
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, err
	}

	var url model.Url
	utils.Copy(&url, &req)

	url.PopulateValues()
	url.ShortId = utils.IdGenerator(config.GetConfig().ShortIdLength.Default)

	if url.UserId = &id; id == "" {
		url.UserId = nil
	}

	err := s.repo.Create(ctx, &url)
	if err != nil {
		logger.Errorf("Create.Create failed, long_url: %s, error: %s", url.LongUrl, err)
		return nil, err
	}

	return &url, nil
}
