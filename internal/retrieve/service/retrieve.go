package service

import (
	"context"
	"errors"

	"shortbin/internal/retrieve/repository"
	"shortbin/pkg/config"
)

//go:generate mockery --name=IRetrieveService
type IRetrieveService interface {
	GetLongUrlByShortId(ctx context.Context, shortId string) (string, error)
}

type RetrieveService struct {
	repo repository.IRetrieveRepository
}

func NewRetrieveService(
	repo repository.IRetrieveRepository) *RetrieveService {
	return &RetrieveService{
		repo: repo,
	}
}

func (s RetrieveService) GetLongUrlByShortId(ctx context.Context, shortId string) (string, error) {
	cfg := config.GetConfig()
	if len(shortId) != cfg.ShortIdLength {
		return "", errors.New("length not in range")
	}

	url, err := s.repo.GetUrlByID(ctx, shortId)
	if err != nil {
		return "", err
	}

	longUrl := url.Long
	return longUrl, nil
}
