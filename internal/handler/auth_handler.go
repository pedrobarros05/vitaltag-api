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

// POST /api/v1/auth/registro
func (h *AuthHandler) Registro(w http.ResponseWriter, r *http.Request) {
	var req model.RegistroRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responderJSON(w, http.StatusBadRequest, model.ErroResponse{Erro: "corpo da requisição inválido"})
		return
	}

	resp, err := h.authSvc.Registrar(r.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrConsentimentoObrigatorio):
			responderJSON(w, http.StatusUnprocessableEntity, model.ErroResponse{Erro: err.Error()})
		case errors.Is(err, repository.ErrEmailDuplicado):
			responderJSON(w, http.StatusConflict, model.ErroResponse{Erro: "e-mail já cadastrado"})
		default:
			responderJSON(w, http.StatusInternalServerError, model.ErroResponse{Erro: "erro interno"})
		}
		return
	}

	responderJSON(w, http.StatusCreated, resp)
}

// POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responderJSON(w, http.StatusBadRequest, model.ErroResponse{Erro: "corpo da requisição inválido"})
		return
	}

	resp, err := h.authSvc.Login(r.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrCredenciaisInvalidas) {
			responderJSON(w, http.StatusUnauthorized, model.ErroResponse{Erro: "e-mail ou senha inválidos"})
			return
		}
		responderJSON(w, http.StatusInternalServerError, model.ErroResponse{Erro: "erro interno"})
		return
	}

	responderJSON(w, http.StatusOK, resp)
}

// ---------- helper -----------------------------------------------------------

func responderJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
