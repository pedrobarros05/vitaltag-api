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

type ProfileService struct {
	profileRepo *repository.ProfileRepo
	userRepo    *repository.UserRepo
}

func NewProfileService(pr *repository.ProfileRepo, ur *repository.UserRepo) *ProfileService {
	return &ProfileService{profileRepo: pr, userRepo: ur}
}

// Get returns the complete profile of the authenticated user (RF11).
func (s *ProfileService) Get(ctx context.Context, userID uuid.UUID) (*model.FullProfile, error) {
	return s.profileRepo.FindByUser(ctx, userID)
}

// Save creates or updates the clinical profile (RF04, RF05, RF12).
// The public password is different from the account password (RF05).
func (s *ProfileService) Save(ctx context.Context, userID uuid.UUID, req *model.ProfileRequest) (*model.FullProfile, error) {
	if req.FirstName == "" || req.LastName == "" {
		return nil, errors.New("first name and last name are required")
	}
	if req.PublicPassword == "" {
		return nil, errors.New("public password is required")
	}
	if len(req.PublicPassword) < 4 {
		return nil, errors.New("public password must be at least 4 characters long")
	}

	// Verify that the public password differs from the account password (RF05)
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.PublicPassword)) == nil {
		return nil, errors.New("public password must differ from the account password")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.PublicPassword), bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("hash public password: %w", err)
	}

	return s.profileRepo.Upsert(ctx, userID, req, string(hash))
}

// Delete removes the account and all cascading data (RF10 + LGPD).
func (s *ProfileService) Delete(ctx context.Context, userID uuid.UUID) error {
	return s.userRepo.Delete(ctx, userID)
}
