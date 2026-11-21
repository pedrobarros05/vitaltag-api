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

const maxAttempts = 5
const lockoutDuration = 15 * time.Minute

var ErrProfileLocked ProfileLockedError
var ErrInvalidPublicPassword = errors.New("incorrect public password")

// ProfileLockedError is returned when the profile is temporarily locked (RF09).
type ProfileLockedError struct{}

func (ProfileLockedError) Error() string { return "profile temporarily locked" }

type PublicService struct {
	profileRepo *repository.ProfileRepo
}

func NewPublicService(pr *repository.ProfileRepo) *PublicService {
	return &PublicService{profileRepo: pr}
}

// AccessProfile validates the public password and returns clinical data (RF08, RF09).
func (s *PublicService) AccessProfile(ctx context.Context, slugStr, password, ip string) (*model.FullProfile, error) {
	slug, err := uuid.Parse(slugStr)
	if err != nil {
		return nil, repository.ErrNotFound
	}

	profile, err := s.profileRepo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	// Check for an active lockout (RF09)
	if profile.LockedUntil != nil && profile.LockedUntil.After(time.Now()) {
		return nil, ProfileLockedError{}
	}

	// Validate the public password
	if err := bcrypt.CompareHashAndPassword([]byte(profile.PublicPasswordHash), []byte(password)); err != nil {
		s.recordFailureAndLockIfNeeded(ctx, profile, ip)
		return nil, ErrInvalidPublicPassword
	}

	// Successful access — record it and clear the lockout
	_ = s.profileRepo.RecordAttempt(ctx, profile.ID, ip, true)

	// Clear the hash before returning (do not expose it to the responder)
	profile.PublicPasswordHash = ""
	return profile, nil
}

func (s *PublicService) recordFailureAndLockIfNeeded(ctx context.Context, profile *model.FullProfile, ip string) {
	_ = s.profileRepo.RecordAttempt(ctx, profile.ID, ip, false)

	count, err := s.profileRepo.CountRecentFailures(ctx, profile.ID)
	if err != nil {
		return
	}

	if count >= maxAttempts {
		until := time.Now().Add(lockoutDuration)
		_ = s.profileRepo.Lock(ctx, profile.ID, until)
		fmt.Printf("[security] profile %s locked until %s after %d attempts\n", profile.ID, until, count)
	}
}
