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

type PublicoHandler struct {
	publicoSvc *service.PublicoService
}

func NewPublicoHandler(svc *service.PublicoService) *PublicoHandler {
	return &PublicoHandler{publicoSvc: svc}
}

// POST /api/v1/publico/{slug} — valida senha e retorna dados clínicos (RF08, RF09)
func (h *PublicoHandler) AcessarPerfil(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		responderJSON(w, http.StatusBadRequest, model.ErroResponse{Erro: "slug ausente"})
		return
	}

	var req model.AcessoPublicoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Senha == "" {
		responderJSON(w, http.StatusBadRequest, model.ErroResponse{Erro: "senha obrigatória"})
		return
	}

	ip := extrairIP(r)
	perfil, err := h.publicoSvc.AcessarPerfil(r.Context(), slug, req.Senha, ip)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNaoEncontrado):
			responderJSON(w, http.StatusNotFound, model.ErroResponse{Erro: "perfil não encontrado"})
		case errors.As(err, &service.ErrPerfil):
			responderJSON(w, http.StatusTooManyRequests, model.ErroResponse{Erro: "muitas tentativas — tente novamente em 15 minutos"})
		case errors.Is(err, service.ErrSenhaPublicaInvalida):
			responderJSON(w, http.StatusUnauthorized, model.ErroResponse{Erro: "senha incorreta"})
		default:
			responderJSON(w, http.StatusInternalServerError, model.ErroResponse{Erro: "erro interno"})
		}
		return
	}

	responderJSON(w, http.StatusOK, perfil)
}

func extrairIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.SplitN(xff, ",", 2)[0]
	}
	return r.RemoteAddr
}
