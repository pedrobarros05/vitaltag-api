package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pedrobarros/vitaltag-api/internal/model"
	"github.com/pedrobarros/vitaltag-api/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

const maxTentativas = 5
const bloqueioDuracao = 15 * time.Minute

var ErrPerfil BloqueadoErr
var ErrSenhaPublicaInvalida = errors.New("senha pública incorreta")

// BloqueadoErr é retornado quando o perfil está temporariamente bloqueado (RF09).
type BloqueadoErr struct{}

func (BloqueadoErr) Error() string { return "perfil bloqueado temporariamente" }

type PublicoService struct {
	perfilRepo *repository.PerfilRepo
}

func NewPublicoService(pr *repository.PerfilRepo) *PublicoService {
	return &PublicoService{perfilRepo: pr}
}

// AcessarPerfil valida a senha pública e retorna os dados clínicos (RF08, RF09).
func (s *PublicoService) AcessarPerfil(ctx context.Context, slugStr, senha, ip string) (*model.PerfilCompleto, error) {
	slug, err := uuid.Parse(slugStr)
	if err != nil {
		return nil, repository.ErrNaoEncontrado
	}

	perfil, err := s.perfilRepo.BuscarPorSlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	// Verifica bloqueio ativo (RF09)
	if perfil.BloqueadoAte != nil && perfil.BloqueadoAte.After(time.Now()) {
		return nil, BloqueadoErr{}
	}

	// Valida senha pública
	if err := bcrypt.CompareHashAndPassword([]byte(perfil.SenhaPublicaHash), []byte(senha)); err != nil {
		s.registrarFalhaEBloquearSeNecessario(ctx, perfil, ip)
		return nil, ErrSenhaPublicaInvalida
	}

	// Acesso bem-sucedido — registra e zera o bloqueio
	_ = s.perfilRepo.RegistrarTentativa(ctx, perfil.ID, ip, true)

	// Limpa o hash antes de retornar (não expor ao socorrista)
	perfil.SenhaPublicaHash = ""
	return perfil, nil
}

func (s *PublicoService) registrarFalhaEBloquearSeNecessario(ctx context.Context, perfil *model.PerfilCompleto, ip string) {
	_ = s.perfilRepo.RegistrarTentativa(ctx, perfil.ID, ip, false)

	count, err := s.perfilRepo.ContarFalhasRecentes(ctx, perfil.ID)
	if err != nil {
		return
	}

	if count >= maxTentativas {
		ate := time.Now().Add(bloqueioDuracao)
		_ = s.perfilRepo.Bloquear(ctx, perfil.ID, ate)
		fmt.Printf("[segurança] perfil %s bloqueado até %s após %d tentativas\n", perfil.ID, ate, count)
	}
}
