package http

import (
	"encoding/json"
	"log"
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

	shortURL, err := h.ctrl.GetShortenUrl(r.Context(), req.LongURL)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Error generating short URL", http.StatusInternalServerError)
		return
	}
	err = writeJSON(w, http.StatusCreated, envelope{"shortened_url": shortURL}, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
