package repository

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.elastic.co/apm/v2"

	"shortbin/internal/auth/model"
)

type IUserRepository interface {
	Create(ctx *gin.Context, email string, hashedPassword string) (*model.User, error)
	Update(ctx *gin.Context, user *model.User) error
	UpdatePassword(ctx *gin.Context, userID string, hashedPassword string) error
	GetUserByID(ctx *gin.Context, userID string) (*model.User, error)
	GetUserByEmail(ctx *gin.Context, email string) (*model.User, error)
}

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx *gin.Context, email string, hashedPassword string) (*model.User, error) {
	apmTx := apm.TransactionFromContext(ctx.Request.Context())
	rootSpan := apmTx.StartSpan("*UserRepo.Create", "repository", nil)
	defer rootSpan.End()

	query := `INSERT INTO users (email, hashed_password) VALUES ($1, $2) RETURNING id, created_at, email, hashed_password`

	var createdUser model.User
	if err := r.db.QueryRow(ctx, query, email, hashedPassword).Scan(&createdUser.ID, &createdUser.CreatedAt, &createdUser.Email, &createdUser.HashedPassword); err != nil {
		return nil, err
	}

	return &createdUser, nil
}

func (r *UserRepo) Update(ctx *gin.Context, user *model.User) error {
	apmTx := apm.TransactionFromContext(ctx.Request.Context())
	rootSpan := apmTx.StartSpan("*UserRepo.Update", "repository", nil)
	defer rootSpan.End()

	query := `UPDATE users SET email=$1, hashed_password=$2 WHERE id=$3`
	_, err := r.db.Exec(ctx, query, user.Email, user.HashedPassword, user.ID)
	return err
}

func (r *UserRepo) UpdatePassword(ctx *gin.Context, userID string, hashedPassword string) error {
	apmTx := apm.TransactionFromContext(ctx.Request.Context())
	rootSpan := apmTx.StartSpan("*UserRepo.UpdatePassword", "repository", nil)
	defer rootSpan.End()

	query := `UPDATE users SET hashed_password=$1 WHERE id=$2`
	_, err := r.db.Exec(ctx, query, hashedPassword, userID)
	return err
}

func (r *UserRepo) GetUserByID(ctx *gin.Context, id string) (*model.User, error) {
	apmTx := apm.TransactionFromContext(ctx.Request.Context())
	rootSpan := apmTx.StartSpan("*UserRepo.GetUserByID", "repository", nil)
	defer rootSpan.End()

	query := `SELECT id, created_at, email, hashed_password FROM users WHERE id=$1`

	var user model.User
	if err := r.db.QueryRow(ctx, query, id).Scan(&user.ID, &user.CreatedAt, &user.Email, &user.HashedPassword); err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepo) GetUserByEmail(ctx *gin.Context, email string) (*model.User, error) {
	apmTx := apm.TransactionFromContext(ctx.Request.Context())
	rootSpan := apmTx.StartSpan("*UserRepo.GetUserByEmail", "repository", nil)
	defer rootSpan.End()

	query := `SELECT id, created_at, email, hashed_password FROM users WHERE email=$1`

	var user model.User
	if err := r.db.QueryRow(ctx, query, email).Scan(&user.ID, &user.CreatedAt, &user.Email, &user.HashedPassword); err != nil {
		return nil, err
	}

	return &user, nil
}
