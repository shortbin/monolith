package repository

import (
	"errors"

	"github.com/gin-gonic/gin"
	"go.elastic.co/apm/v2"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"shortbin/internal/common/model"
)

type IRetrieveRepository interface {
	GetUrlByID(ctx *gin.Context, id string) (*model.Url, error)
}

type RetrieveRepo struct {
	db *pgxpool.Pool
}

func NewRetrieveRepository(db *pgxpool.Pool) *RetrieveRepo {
	return &RetrieveRepo{db: db}
}

func (r *RetrieveRepo) GetUrlByID(ctx *gin.Context, id string) (*model.Url, error) {
	apmTx := apm.TransactionFromContext(ctx.Request.Context())
	rootSpan := apmTx.StartSpan("*RetrieveRepo.GetUrlByID", "repository", nil)
	defer rootSpan.End()

	query := `SELECT short_id, long_url, user_id, created_at, expires_at FROM urls WHERE short_id=$1`

	row := r.db.QueryRow(ctx, query, id)

	var url model.Url
	if err := row.Scan(&url.ShortId, &url.LongUrl, &url.UserId, &url.CreatedAt, &url.ExpiresAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("id not found")
		}
		return nil, err
	}

	return &url, nil
}
