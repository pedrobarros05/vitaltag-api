package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/pedrobarros/vitaltag-api/internal/model"
	"github.com/pedrobarros/vitaltag-api/internal/repository"
	"github.com/pedrobarros/vitaltag-api/internal/service"
)

type PublicHandler struct {
	publicSvc *service.PublicService
}

func NewPublicHandler(svc *service.PublicService) *PublicHandler {
	return &PublicHandler{publicSvc: svc}
}

// POST /api/v1/public/{slug} — validates the password and returns clinical data (RF08, RF09)
func (h *PublicHandler) AccessProfile(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		respondJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "missing slug"})
		return
	}

	var req model.PublicAccessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Password == "" {
		respondJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "password is required"})
		return
	}

	ip := extractIP(r)
	profile, err := h.publicSvc.AccessProfile(r.Context(), slug, req.Password, ip)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			respondJSON(w, http.StatusNotFound, model.ErrorResponse{Error: "profile not found"})
		case errors.As(err, &service.ErrProfileLocked):
			respondJSON(w, http.StatusTooManyRequests, model.ErrorResponse{Error: "too many attempts — try again in 15 minutes"})
		case errors.Is(err, service.ErrInvalidPublicPassword):
			respondJSON(w, http.StatusUnauthorized, model.ErrorResponse{Error: "incorrect password"})
		default:
			respondJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "internal error"})
		}
		return
	}

	respondJSON(w, http.StatusOK, profile)
}

func extractIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.SplitN(xff, ",", 2)[0]
	}
	return r.RemoteAddr
}
