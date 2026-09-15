package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/pedrobarros/vitaltag-api/internal/model"
	"github.com/pedrobarros/vitaltag-api/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

const bcryptCusto = 12

var ErrCredenciaisInvalidas = errors.New("credenciais inválidas")
var ErrConsentimentoObrigatorio = errors.New("consentimento LGPD é obrigatório")

type AuthService struct {
	usuarioRepo *repository.UsuarioRepo
}

func NewAuthService(r *repository.UsuarioRepo) *AuthService {
	return &AuthService{usuarioRepo: r}
}

// Registrar cria conta (RF01). Requer consentimento LGPD (RNF04).
func (s *AuthService) Registrar(ctx context.Context, req *model.RegistroRequest) (*model.TokenResponse, error) {
	if !req.Consentimento {
		return nil, ErrConsentimentoObrigatorio
	}
	if len(req.Senha) < 8 {
		return nil, errors.New("senha deve ter no mínimo 8 caracteres")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Senha), bcryptCusto)
	if err != nil {
		return nil, fmt.Errorf("hash de senha: %w", err)
	}

	u, err := s.usuarioRepo.Criar(ctx, req.Email, string(hash), req.Consentimento)
	if err != nil {
		if errors.Is(err, repository.ErrEmailDuplicado) {
			return nil, repository.ErrEmailDuplicado
		}
		return nil, fmt.Errorf("criar usuário: %w", err)
	}

	token, err := gerarJWT(u.ID)
	if err != nil {
		return nil, err
	}
	return &model.TokenResponse{Token: token}, nil
}

// Login autentica o usuário e devolve um JWT (RF02).
func (s *AuthService) Login(ctx context.Context, req *model.LoginRequest) (*model.TokenResponse, error) {
	u, err := s.usuarioRepo.BuscarPorEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNaoEncontrado) {
			return nil, ErrCredenciaisInvalidas
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.SenhaHash), []byte(req.Senha)); err != nil {
		return nil, ErrCredenciaisInvalidas
	}

	token, err := gerarJWT(u.ID)
	if err != nil {
		return nil, err
	}
	return &model.TokenResponse{Token: token}, nil
}

// ---------- JWT ---------------------------------------------------------------

type Claims struct {
	UsuarioID uuid.UUID `json:"uid"`
	jwt.RegisteredClaims
}

func gerarJWT(usuarioID uuid.UUID) (string, error) {
	segredo := []byte(jwtSegredo())
	claims := Claims{
		UsuarioID: usuarioID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(segredo)
}

func ValidarJWT(tokenStr string) (*Claims, error) {
	segredo := []byte(jwtSegredo())
	t, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de assinatura inesperado: %v", t.Header["alg"])
		}
		return segredo, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := t.Claims.(*Claims)
	if !ok || !t.Valid {
		return nil, errors.New("token inválido")
	}
	return claims, nil
}

func jwtSegredo() string {
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		s = "dev-secret-troque-em-producao"
	}
	return s
}
