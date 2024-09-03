package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"shortbin/internal/auth/model"
)

type IUserRepository interface {
	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, user *model.User) error
	GetUserByID(ctx context.Context, id string) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
}

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, user *model.User) error {

	query := `INSERT INTO users (id, created_at, email, password) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(ctx, query, user.ID, user.CreatedAt, user.Email, user.Password)
	return err
}

func (r *UserRepo) Update(ctx context.Context, user *model.User) error {

	query := `UPDATE users SET email=$1, password=$2 WHERE id=$3`
	_, err := r.db.Exec(ctx, query, user.Email, user.Password, user.ID)
	return err
}

func (r *UserRepo) GetUserByID(ctx context.Context, id string) (*model.User, error) {

	query := `SELECT id, created_at, email, password FROM users WHERE id=$1`
	row := r.db.QueryRow(ctx, query, id)

	var user model.User
	if err := row.Scan(&user.ID, &user.CreatedAt, &user.Email, &user.Password); err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepo) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {

	query := `SELECT id, created_at, email, password FROM users WHERE email=$1`
	row := r.db.QueryRow(ctx, query, email)

	var user model.User
	if err := row.Scan(&user.ID, &user.CreatedAt, &user.Email, &user.Password); err != nil {
		return nil, err
	}

	return &user, nil
}
