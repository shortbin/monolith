package service

import (
	"errors"

	"github.com/gin-gonic/gin"
	"go.elastic.co/apm/v2"

	"shortbin/internal/retrieve/repository"
	"shortbin/pkg/config"
	"shortbin/pkg/response"
)

//go:generate mockery --name=IRetrieveService
type IRetrieveService interface {
	Retrieve(ctx *gin.Context, shortID string) (string, error)
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

func (s *RetrieveService) Retrieve(ctx *gin.Context, shortID string) (string, error) {
	apmTx := apm.TransactionFromContext(ctx.Request.Context())
	rootSpan := apmTx.StartSpan("*RetrieveService.Retrieve", "service", nil)
	defer rootSpan.End()

	cfg := config.GetConfig()

	if length := len(shortID); length < cfg.ShortIDLength.Min || cfg.ShortIDLength.Max < length {
		return "", errors.New(response.IDLengthNotInRange)
	}

	url, err := s.repo.GetURLByID(ctx, shortID)
	if err != nil {
		return "", err
	}

	longURL := url.LongURL
	return longURL, nil
}
