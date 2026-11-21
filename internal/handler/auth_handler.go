package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/pedrobarros/vitaltag-api/internal/model"
	"github.com/pedrobarros/vitaltag-api/internal/repository"
	"github.com/pedrobarros/vitaltag-api/internal/service"
)

type AuthHandler struct {
	authSvc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: svc}
}

// POST /api/v1/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "invalid request body"})
		return
	}

	resp, err := h.authSvc.Register(r.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrConsentRequired):
			respondJSON(w, http.StatusUnprocessableEntity, model.ErrorResponse{Error: err.Error()})
		case errors.Is(err, repository.ErrDuplicateEmail):
			respondJSON(w, http.StatusConflict, model.ErrorResponse{Error: "email already registered"})
		default:
			respondJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "internal error"})
		}
		return
	}

	respondJSON(w, http.StatusCreated, resp)
}

// POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "invalid request body"})
		return
	}

	resp, err := h.authSvc.Login(r.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			respondJSON(w, http.StatusUnauthorized, model.ErrorResponse{Error: "invalid email or password"})
			return
		}
		respondJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "internal error"})
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

// ---------- helper -----------------------------------------------------------

func respondJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
