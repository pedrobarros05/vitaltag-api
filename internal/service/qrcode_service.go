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
	perfilRepo *repository.PerfilRepo
}

func NewQRCodeService(pr *repository.PerfilRepo) *QRCodeService {
	return &QRCodeService{perfilRepo: pr}
}

// Gerar gera o QR Code do perfil do usuário em PNG (RF06).
// O QR Code contém apenas a URL pública — sem dados clínicos.
func (s *QRCodeService) Gerar(ctx context.Context, usuarioID uuid.UUID) ([]byte, error) {
	perfil, err := s.perfilRepo.BuscarPorUsuario(ctx, usuarioID)
	if err != nil {
		if err == repository.ErrNaoEncontrado {
			return nil, fmt.Errorf("crie seu perfil clínico antes de gerar o QR Code")
		}
		return nil, fmt.Errorf("buscar perfil: %w", err)
	}

	url := fmt.Sprintf("%s/publico/%s", baseURL(), perfil.SlugPublico.String())

	var buf bytes.Buffer
	png, err := qrcode.Encode(url, qrcode.High, 256)
	if err != nil {
		return nil, fmt.Errorf("gerar QR Code: %w", err)
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
