package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pedrobarros/vitaltag-api/internal/model"
)

var ErrNotFound = errors.New("not found")
var ErrDuplicateEmail = errors.New("email already registered")

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

// Create inserts a new user. The password must already be a bcrypt hash.
func (r *UserRepo) Create(ctx context.Context, email, passwordHash string, consent bool) (*model.User, error) {
	u := &model.User{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, consent)
		 VALUES ($1, $2, $3)
		 RETURNING id, email, password_hash, consent, created_at, updated_at`,
		email, passwordHash, consent,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Consent, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if isPgUniqueViolation(err) {
			return nil, ErrDuplicateEmail
		}
		return nil, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}

// FindByEmail returns the user with password_hash (for verification).
func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	u := &model.User{}
	err := r.db.QueryRow(ctx,
		`SELECT id, email, password_hash, consent, created_at, updated_at
		 FROM users WHERE email = $1`,
		email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Consent, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find user: %w", err)
	}
	return u, nil
}

// FindByID returns a user by UUID.
func (r *UserRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	u := &model.User{}
	err := r.db.QueryRow(ctx,
		`SELECT id, email, password_hash, consent, created_at, updated_at
		 FROM users WHERE id = $1`,
		id,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Consent, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return u, nil
}

// Delete removes the user and cascades to the profile and clinical data
// through ON DELETE CASCADE.
func (r *UserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	ct, err := r.db.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ---------- helpers ----------------------------------------------------------

func isPgUniqueViolation(err error) bool {
	return err != nil && (pgErrCode(err) == "23505")
}

func pgErrCode(err error) string {
	type pgError interface{ SQLState() string }
	if pe, ok := err.(pgError); ok {
		return pe.SQLState()
	}
	return ""
}
