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

var ErrNaoEncontrado = errors.New("não encontrado")
var ErrEmailDuplicado = errors.New("e-mail já cadastrado")

type UsuarioRepo struct {
	db *pgxpool.Pool
}

func NewUsuarioRepo(db *pgxpool.Pool) *UsuarioRepo {
	return &UsuarioRepo{db: db}
}

// Criar inserts a new user. The password must already be a bcrypt hash.
func (r *UsuarioRepo) Criar(ctx context.Context, email, senhaHash string, consentimento bool) (*model.Usuario, error) {
	u := &model.Usuario{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO usuario (email, senha_hash, consentimento)
		 VALUES ($1, $2, $3)
		 RETURNING id, email, senha_hash, consentimento, criado_em, atualizado_em`,
		email, senhaHash, consentimento,
	).Scan(&u.ID, &u.Email, &u.SenhaHash, &u.Consentimento, &u.CriadoEm, &u.AtualizadoEm)
	if err != nil {
		if isPgUniqueViolation(err) {
			return nil, ErrEmailDuplicado
		}
		return nil, fmt.Errorf("criar usuario: %w", err)
	}
	return u, nil
}

// BuscarPorEmail returns the user with senha_hash (for verification).
func (r *UsuarioRepo) BuscarPorEmail(ctx context.Context, email string) (*model.Usuario, error) {
	u := &model.Usuario{}
	err := r.db.QueryRow(ctx,
		`SELECT id, email, senha_hash, consentimento, criado_em, atualizado_em
		 FROM usuario WHERE email = $1`,
		email,
	).Scan(&u.ID, &u.Email, &u.SenhaHash, &u.Consentimento, &u.CriadoEm, &u.AtualizadoEm)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNaoEncontrado
		}
		return nil, fmt.Errorf("buscar usuario: %w", err)
	}
	return u, nil
}

// BuscarPorID returns a user by UUID.
func (r *UsuarioRepo) BuscarPorID(ctx context.Context, id uuid.UUID) (*model.Usuario, error) {
	u := &model.Usuario{}
	err := r.db.QueryRow(ctx,
		`SELECT id, email, senha_hash, consentimento, criado_em, atualizado_em
		 FROM usuario WHERE id = $1`,
		id,
	).Scan(&u.ID, &u.Email, &u.SenhaHash, &u.Consentimento, &u.CriadoEm, &u.AtualizadoEm)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNaoEncontrado
		}
		return nil, fmt.Errorf("buscar usuario por id: %w", err)
	}
	return u, nil
}

// Excluir removes the user and cascades to the profile and clinical data
// through ON DELETE CASCADE.
func (r *UsuarioRepo) Excluir(ctx context.Context, id uuid.UUID) error {
	ct, err := r.db.Exec(ctx, `DELETE FROM usuario WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("excluir usuario: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrNaoEncontrado
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
