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

type ProfileHandler struct {
	profileSvc *service.ProfileService
	qrCodeSvc  *service.QRCodeService
}

func NewProfileHandler(ps *service.ProfileService, qs *service.QRCodeService) *ProfileHandler {
	return &ProfileHandler{profileSvc: ps, qrCodeSvc: qs}
}

// GET /api/v1/profile — returns the authenticated user's profile (RF11)
func (h *ProfileHandler) Get(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respondJSON(w, http.StatusUnauthorized, model.ErrorResponse{Error: "unauthorized"})
		return
	}

	profile, err := h.profileSvc.Get(r.Context(), uid)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respondJSON(w, http.StatusNotFound, model.ErrorResponse{Error: "profile not found — create one first"})
			return
		}
		respondJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "failed to fetch profile"})
		return
	}

	// Do not expose the public-password hash
	profile.PublicPasswordHash = ""
	respondJSON(w, http.StatusOK, profile)
}

// POST /api/v1/profile — creates or updates the clinical profile (RF04, RF05, RF12)
func (h *ProfileHandler) Save(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respondJSON(w, http.StatusUnauthorized, model.ErrorResponse{Error: "unauthorized"})
		return
	}

	var req model.ProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "invalid request body"})
		return
	}

	profile, err := h.profileSvc.Save(r.Context(), uid, &req)
	if err != nil {
		respondJSON(w, http.StatusUnprocessableEntity, model.ErrorResponse{Error: err.Error()})
		return
	}

	profile.PublicPasswordHash = ""
	respondJSON(w, http.StatusOK, profile)
}

// GET /api/v1/profile/qrcode — returns the QR Code as a PNG (RF06)
func (h *ProfileHandler) QRCode(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respondJSON(w, http.StatusUnauthorized, model.ErrorResponse{Error: "unauthorized"})
		return
	}

	png, err := h.qrCodeSvc.Generate(r.Context(), uid)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(http.StatusOK)
	w.Write(png)
}

// DELETE /api/v1/profile — deletes the account and all data (RF10)
func (h *ProfileHandler) Delete(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respondJSON(w, http.StatusUnauthorized, model.ErrorResponse{Error: "unauthorized"})
		return
	}

	if err := h.profileSvc.Delete(r.Context(), uid); err != nil {
		respondJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "failed to delete account"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
