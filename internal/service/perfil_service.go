package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/pedrobarros/vitaltag-api/internal/model"
	"github.com/pedrobarros/vitaltag-api/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type PerfilService struct {
	perfilRepo  *repository.PerfilRepo
	usuarioRepo *repository.UsuarioRepo
}

func NewPerfilService(pr *repository.PerfilRepo, ur *repository.UsuarioRepo) *PerfilService {
	return &PerfilService{perfilRepo: pr, usuarioRepo: ur}
}

// Buscar returns the complete profile of the authenticated user (RF11).
func (s *PerfilService) Buscar(ctx context.Context, usuarioID uuid.UUID) (*model.PerfilCompleto, error) {
	return s.perfilRepo.BuscarPorUsuario(ctx, usuarioID)
}

// Salvar creates or updates the clinical profile (RF04, RF05, RF12).
// The public password is different from the account password (RF05).
func (s *PerfilService) Salvar(ctx context.Context, usuarioID uuid.UUID, req *model.PerfilRequest) (*model.PerfilCompleto, error) {
	if req.Nome == "" || req.Sobrenome == "" {
		return nil, errors.New("nome e sobrenome são obrigatórios")
	}
	if req.SenhaPublica == "" {
		return nil, errors.New("senha pública é obrigatória")
	}
	if len(req.SenhaPublica) < 4 {
		return nil, errors.New("senha pública deve ter no mínimo 4 caracteres")
	}

	// Verify that the public password differs from the account password (RF05)
	u, err := s.usuarioRepo.BuscarPorID(ctx, usuarioID)
	if err != nil {
		return nil, fmt.Errorf("buscar usuário: %w", err)
	}
	if bcrypt.CompareHashAndPassword([]byte(u.SenhaHash), []byte(req.SenhaPublica)) == nil {
		return nil, errors.New("a senha pública não pode ser igual à senha da conta")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.SenhaPublica), bcryptCusto)
	if err != nil {
		return nil, fmt.Errorf("hash senha pública: %w", err)
	}

	return s.perfilRepo.SalvarOuAtualizar(ctx, usuarioID, req, string(hash))
}

// Excluir removes the account and all cascading data (RF10 + LGPD).
func (s *PerfilService) Excluir(ctx context.Context, usuarioID uuid.UUID) error {
	return s.usuarioRepo.Excluir(ctx, usuarioID)
}
