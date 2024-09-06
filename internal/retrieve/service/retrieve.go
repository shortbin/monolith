package service

import (
	"context"
	"errors"

	"shortbin/internal/retrieve/repository"
	"shortbin/pkg/config"
)

//go:generate mockery --name=IRetrieveService
type IRetrieveService interface {
	Retrieve(ctx context.Context, shortId string) (string, error)
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

func (s RetrieveService) Retrieve(ctx context.Context, shortId string) (string, error) {
	cfg := config.GetConfig()

	if length := len(shortId); length < cfg.ShortIdLength.Min || cfg.ShortIdLength.Max < length {
		return "", errors.New("length not in range")
	}

	url, err := s.repo.GetUrlByID(ctx, shortId)
	if err != nil {
		return "", err
	}

	longUrl := url.LongUrl
	return longUrl, nil
}
