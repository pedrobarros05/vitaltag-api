package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/pedrobarros/vitaltag-api/internal/middleware"
	"github.com/pedrobarros/vitaltag-api/internal/model"
	"github.com/pedrobarros/vitaltag-api/internal/repository"
	"github.com/pedrobarros/vitaltag-api/internal/service"
)

type PerfilHandler struct {
	perfilSvc  *service.PerfilService
	qrCodeSvc  *service.QRCodeService
}

func NewPerfilHandler(ps *service.PerfilService, qs *service.QRCodeService) *PerfilHandler {
	return &PerfilHandler{perfilSvc: ps, qrCodeSvc: qs}
}

// GET /api/v1/perfil — retorna o perfil do usuário autenticado (RF11)
func (h *PerfilHandler) Buscar(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UsuarioIDDoContexto(r.Context())
	if !ok {
		responderJSON(w, http.StatusUnauthorized, model.ErroResponse{Erro: "não autorizado"})
		return
	}

	perfil, err := h.perfilSvc.Buscar(r.Context(), uid)
	if err != nil {
		if errors.Is(err, repository.ErrNaoEncontrado) {
			responderJSON(w, http.StatusNotFound, model.ErroResponse{Erro: "perfil não encontrado — crie um primeiro"})
			return
		}
		responderJSON(w, http.StatusInternalServerError, model.ErroResponse{Erro: "erro ao buscar perfil"})
		return
	}

	// Não expor o hash da senha pública
	perfil.SenhaPublicaHash = ""
	responderJSON(w, http.StatusOK, perfil)
}

// POST /api/v1/perfil — cria/atualiza perfil clínico (RF04, RF05, RF12)
func (h *PerfilHandler) SalvarOuAtualizar(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UsuarioIDDoContexto(r.Context())
	if !ok {
		responderJSON(w, http.StatusUnauthorized, model.ErroResponse{Erro: "não autorizado"})
		return
	}

	var req model.PerfilRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responderJSON(w, http.StatusBadRequest, model.ErroResponse{Erro: "corpo da requisição inválido"})
		return
	}

	perfil, err := h.perfilSvc.Salvar(r.Context(), uid, &req)
	if err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, model.ErroResponse{Erro: err.Error()})
		return
	}

	perfil.SenhaPublicaHash = ""
	responderJSON(w, http.StatusOK, perfil)
}

// GET /api/v1/perfil/qrcode — devolve QR Code em PNG (RF06)
func (h *PerfilHandler) QRCode(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UsuarioIDDoContexto(r.Context())
	if !ok {
		responderJSON(w, http.StatusUnauthorized, model.ErroResponse{Erro: "não autorizado"})
		return
	}

	png, err := h.qrCodeSvc.Gerar(r.Context(), uid)
	if err != nil {
		responderJSON(w, http.StatusBadRequest, model.ErroResponse{Erro: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(http.StatusOK)
	w.Write(png)
}

// DELETE /api/v1/perfil — exclui conta e todos os dados (RF10)
func (h *PerfilHandler) Excluir(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UsuarioIDDoContexto(r.Context())
	if !ok {
		responderJSON(w, http.StatusUnauthorized, model.ErroResponse{Erro: "não autorizado"})
		return
	}

	if err := h.perfilSvc.Excluir(r.Context(), uid); err != nil {
		responderJSON(w, http.StatusInternalServerError, model.ErroResponse{Erro: "erro ao excluir conta"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
