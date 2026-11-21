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

const bcryptCost = 12

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrConsentRequired = errors.New("LGPD consent is required")

type AuthService struct {
	userRepo *repository.UserRepo
}

func NewAuthService(r *repository.UserRepo) *AuthService {
	return &AuthService{userRepo: r}
}

// Register creates an account (RF01). LGPD consent is required (RNF04).
func (s *AuthService) Register(ctx context.Context, req *model.RegisterRequest) (*model.TokenResponse, error) {
	if !req.Consent {
		return nil, ErrConsentRequired
	}
	if len(req.Password) < 8 {
		return nil, errors.New("password must be at least 8 characters long")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	u, err := s.userRepo.Create(ctx, req.Email, string(hash), req.Consent)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateEmail) {
			return nil, repository.ErrDuplicateEmail
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	token, err := generateJWT(u.ID)
	if err != nil {
		return nil, err
	}
	return &model.TokenResponse{Token: token}, nil
}

// Login authenticates the user and returns a JWT (RF02).
func (s *AuthService) Login(ctx context.Context, req *model.LoginRequest) (*model.TokenResponse, error) {
	u, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := generateJWT(u.ID)
	if err != nil {
		return nil, err
	}
	return &model.TokenResponse{Token: token}, nil
}

// ---------- JWT ---------------------------------------------------------------

type Claims struct {
	UserID uuid.UUID `json:"uid"`
	jwt.RegisteredClaims
}

func generateJWT(userID uuid.UUID) (string, error) {
	secret := []byte(jwtSecret())
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(secret)
}

func ValidateJWT(tokenStr string) (*Claims, error) {
	secret := []byte(jwtSecret())
	t, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := t.Claims.(*Claims)
	if !ok || !t.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func jwtSecret() string {
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		s = "dev-secret-change-in-production"
	}
	return s
}
