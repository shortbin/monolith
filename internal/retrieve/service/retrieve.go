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
	Retrieve(ctx *gin.Context, shortId string) (string, error)
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

func (s *RetrieveService) Retrieve(ctx *gin.Context, shortId string) (string, error) {
	apmTx := apm.TransactionFromContext(ctx.Request.Context())
	rootSpan := apmTx.StartSpan("*RetrieveService.Retrieve", "service", nil)
	defer rootSpan.End()

	cfg := config.GetConfig()

	if length := len(shortId); length < cfg.ShortIdLength.Min || cfg.ShortIdLength.Max < length {
		return "", errors.New(response.IdLengthNotInRange)
	}

	url, err := s.repo.GetUrlByID(ctx, shortId)
	if err != nil {
		return "", err
	}

	longUrl := url.LongUrl
	return longUrl, nil
}
