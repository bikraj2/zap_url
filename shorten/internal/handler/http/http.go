package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bikraj2/url_shortener/shorten/internal/controller"
)

type handler struct {
	ctrl *controller.Controller
}
type envelope map[string]interface{}

func New(ctrl *controller.Controller) *handler {
	return &handler{ctrl}
}

func (h *handler) CreateShortenUrl(w http.ResponseWriter, r *http.Request) {
	var req struct {
		LongURL string `json:"long_url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.LongURL == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	shortURL, err := h.ctrl.CreateShortUrl(r.Context(), req.LongURL)
	if err != nil {
		// log.Println(err.Error())
		fmt.Println(err.Error())
		http.Error(w, fmt.Sprintf("Error generating short URL:%v", err.Error()), http.StatusInternalServerError)
		return
	}
	err = writeJSON(w, http.StatusCreated, envelope{"short_url": shortURL}, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
