package service

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/pedrobarros/vitaltag-api/internal/repository"
	qrcode "github.com/skip2/go-qrcode"
)

type QRCodeService struct {
	profileRepo *repository.ProfileRepo
}

func NewQRCodeService(pr *repository.ProfileRepo) *QRCodeService {
	return &QRCodeService{profileRepo: pr}
}

// Generate generates the user's profile QR Code as a PNG (RF06).
// The QR Code contains only the public URL — no clinical data.
func (s *QRCodeService) Generate(ctx context.Context, userID uuid.UUID) ([]byte, error) {
	profile, err := s.profileRepo.FindByUser(ctx, userID)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, fmt.Errorf("create your clinical profile before generating the QR Code")
		}
		return nil, fmt.Errorf("find profile: %w", err)
	}

	url := fmt.Sprintf("%s/public/%s", baseURL(), profile.PublicSlug.String())

	var buf bytes.Buffer
	png, err := qrcode.Encode(url, qrcode.High, 256)
	if err != nil {
		return nil, fmt.Errorf("generate QR Code: %w", err)
	}
	buf.Write(png)
	return buf.Bytes(), nil
}

func baseURL() string {
	u := os.Getenv("PUBLIC_BASE_URL")
	if u == "" {
		u = "http://localhost:3000"
	}
	return u
}
