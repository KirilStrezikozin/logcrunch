// Copyright 2025 The Logcrunch Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package handlers

import (
	"net/http"

	"github.com/KirilStrezikozin/logcrunch/internal/services"
	"github.com/KirilStrezikozin/logcrunch/web/templates"
	"github.com/a-h/templ"
	"github.com/rs/zerolog"
)

type Handler struct {
	logger      zerolog.Logger
	connService services.IConnectionService
}

func New(logger zerolog.Logger, connService services.IConnectionService) *Handler {
	return &Handler{
		logger:      logger,
		connService: connService,
	}
}

func (h *Handler) Index() http.Handler {
	return templ.Handler(templates.Index())
}

func (h *Handler) Static() http.Handler {
	fs := http.FileServer(http.Dir("web/static"))
	return http.StripPrefix("/static/", fs)
}

func (h *Handler) PostConnectionURL(w http.ResponseWriter, r *http.Request) {
	value := r.FormValue(templates.ConnectionURLInputName)
	value, err := h.connService.SetURL(value)

	if err != nil {
		h.logger.Error().Err(err).Msg("failed to set connection url")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	component := templates.ConnectionURLInput(value, false)
	if err = component.Render(ctx, w); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetConnectionURL(w http.ResponseWriter, r *http.Request) {
	value, err := h.connService.GetURL()
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to get connection url")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	component := templates.ConnectionURLInput(value, false)
	if err = component.Render(ctx, w); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	h.connService.ConnectOnce() // Initial connection attempt.
}

func (h *Handler) GetConnectionStatus(w http.ResponseWriter, r *http.Request) {
	value, err := h.connService.GetStatus()
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to get connection status")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	component := templates.ConnectionStatus(value)
	if err = component.Render(ctx, w); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
