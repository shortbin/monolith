package repository

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.elastic.co/apm/v2"

	"shortbin/internal/common/model"
)

type ICreateRepository interface {
	Create(ctx *gin.Context, url *model.Url) error
}

type CreateRepo struct {
	db *pgxpool.Pool
}

func NewCreateRepository(db *pgxpool.Pool) *CreateRepo {
	return &CreateRepo{db: db}
}

func (r *CreateRepo) Create(ctx *gin.Context, url *model.Url) error {
	apmTx := apm.TransactionFromContext(ctx.Request.Context())
	rootSpan := apmTx.StartSpan("*CreateRepo.Create", "repository", nil)
	defer rootSpan.End()

	query := `INSERT INTO urls (short_id, long_url, user_id, created_at, expires_at) VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.Exec(ctx, query, url.ShortId, url.LongUrl, url.UserId, url.CreatedAt, url.ExpiresAt)
	return err
}
