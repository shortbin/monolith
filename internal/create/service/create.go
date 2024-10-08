package service

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.elastic.co/apm/module/apmzap/v2"
	"go.elastic.co/apm/v2"

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
	Create(ctx *gin.Context, id string, req *dto.CreateReq) (*model.URL, error)
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

func (s *CreateService) Create(ctx *gin.Context, id string, req *dto.CreateReq) (*model.URL, error) {
	apmTx := apm.TransactionFromContext(ctx.Request.Context())
	traceContextFields := apmzap.TraceContext(ctx.Request.Context())
	rootSpan := apmTx.StartSpan("*CreateService.Create", "service", nil)
	defer rootSpan.End()

	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, err
	}

	var url model.URL
	utils.Copy(&url, &req)
	url.CreatedAt = time.Now()
	if url.ExpiresAt.IsZero() {
		url.ExpiresAt = url.CreatedAt.AddDate(
			config.GetConfig().ExpirationInYears,
			0,
			0,
		)
	}

	idGenSpan := apmTx.StartSpan("utils.IdGenerator", "utils", nil)
	url.ShortID = utils.IDGenerator(config.GetConfig().ShortIDLength.Default)
	idGenSpan.End()

	if url.UserID = &id; id == "" {
		url.UserID = nil
	}

	err := s.repo.Create(ctx, &url)
	if err != nil {
		logger.Infof("Create.Create failed, long_url: %s, error: %s", url.LongURL, err)
		logger.ApmLogger.With(traceContextFields...).Error(err.Error())
		return nil, err
	}

	return &url, nil
}
