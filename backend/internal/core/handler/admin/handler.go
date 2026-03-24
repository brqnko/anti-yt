package admin

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/brqnko/anti-yt/backend/internal/admin"
	"github.com/brqnko/anti-yt/backend/internal/core"
)

type handler struct {
	adminService *admin.Service
}

func newHandler(adminService *admin.Service) *handler {
	return &handler{adminService: adminService}
}

func (h *handler) createValuableChannel(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ExternalChannelID string `json:"external_channel_id"`
		Reason            string `json:"reason"`
		Description       string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "Bad Request", "invalid request body")
		return
	}

	vc, err := h.adminService.CreateNewValuableChannel(r.Context(), body.ExternalChannelID, body.Reason, body.Description)
	if err != nil {
		var domainErr *core.DomainError
		if errors.As(err, &domainErr) {
			writeErrorJSON(w, http.StatusBadRequest, domainErr.Code(), domainErr.Error())
			return
		}
		slog.Error("failed to create valuable channel", "error", err)
		writeErrorJSON(w, http.StatusInternalServerError, "Internal Server Error", "an unexpected error has occurred")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(struct {
		ChannelID           string `json:"channel_id"`
		ValuableReasonCode  string `json:"valuable_reason_code"`
		ValuableDescription string `json:"valuable_description"`
	}{
		ChannelID:           vc.ChannelID.String(),
		ValuableReasonCode:  vc.ValuableReasonCode.String(),
		ValuableDescription: vc.ValuableDescription.String(),
	})
}

func (h *handler) updateValuableChannel(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ExternalChannelID string  `json:"external_channel_id"`
		Reason            *string `json:"reason"`
		Description       *string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "Bad Request", "invalid request body")
		return
	}

	vc, err := h.adminService.UpdateValuableChannel(r.Context(), body.ExternalChannelID, body.Reason, body.Description)
	if err != nil {
		var domainErr *core.DomainError
		if errors.As(err, &domainErr) {
			writeErrorJSON(w, http.StatusBadRequest, domainErr.Code(), domainErr.Error())
			return
		}
		slog.Error("failed to update valuable channel", "error", err)
		writeErrorJSON(w, http.StatusInternalServerError, "Internal Server Error", "an unexpected error has occurred")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(struct {
		ChannelID           string `json:"channel_id"`
		ValuableReasonCode  string `json:"valuable_reason_code"`
		ValuableDescription string `json:"valuable_description"`
	}{
		ChannelID:           vc.ChannelID.String(),
		ValuableReasonCode:  vc.ValuableReasonCode.String(),
		ValuableDescription: vc.ValuableDescription.String(),
	})
}

func (h *handler) removeValuableChannel(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ExternalChannelID string `json:"external_channel_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "Bad Request", "invalid request body")
		return
	}

	if err := h.adminService.RemoveValuableChannel(r.Context(), body.ExternalChannelID); err != nil {
		var domainErr *core.DomainError
		if errors.As(err, &domainErr) {
			writeErrorJSON(w, http.StatusBadRequest, domainErr.Code(), domainErr.Error())
			return
		}
		slog.Error("failed to remove valuable channel", "error", err)
		writeErrorJSON(w, http.StatusInternalServerError, "Internal Server Error", "an unexpected error has occurred")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeErrorJSON(w http.ResponseWriter, statusCode int, title, detail string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(struct {
		Title  string `json:"title"`
		Detail string `json:"detail"`
	}{
		Title:  title,
		Detail: detail,
	})
}
