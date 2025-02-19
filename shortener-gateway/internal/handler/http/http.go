package httphandler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	customerror "github.com/bikraj2/url_shortener/shortener-gateway/internal"
	"github.com/bikraj2/url_shortener/shortener-gateway/internal/controller"
	"github.com/bikraj2/url_shortener/shortener-gateway/internal/helper"
	"github.com/go-chi/chi/v5"
)

type handler struct {
	ctrl *controller.Controller
}

func New(ctrl *controller.Controller) *handler {
	return &handler{ctrl}
}

func (h *handler) CreateShortUrl(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var URL struct {
		LongUrl string `json:"long_url"`
	}
	err := json.NewDecoder(r.Body).Decode(&URL)
	if err != nil {
		customerror.ErrorResponse(w, http.StatusInternalServerError, nil, err.Error())
		return
	}
	short_url, err := h.ctrl.CreateShortUrl(r.Context(), URL.LongUrl)
	if err != nil {
		switch {
		case errors.Is(err, customerror.ErrNotFound):
			customerror.NotFoundResponse(w)
		case errors.Is(err, customerror.ErrBadRequest):
			customerror.BadRequestErrorResponse(w)
		default:
			customerror.ErrorResponse(w, http.StatusInternalServerError, nil, err.Error())
		}
		return
	}
	err = helper.WriteJSON(w, http.StatusCreated, helper.Envelope{"short_url": "http://localhost:8084/" + short_url}, nil)
	if err != nil {
		customerror.ErrorResponse(w, http.StatusInternalServerError, nil, err.Error())
		return
	}
}

func (h *handler) GetLongUrl(w http.ResponseWriter, r *http.Request) {
	short_url := chi.URLParam(r, "short_url")

	long_url, err := h.ctrl.GetLongUrl(r.Context(), short_url)
	if err != nil {
		switch {
		case errors.Is(err, customerror.ErrNotFound):
			customerror.NotFoundResponse(w)
		default:
			customerror.ErrorResponse(w, http.StatusInternalServerError, nil, err.Error())
		}
		return
	}
	if !strings.HasPrefix(long_url, "http://") && !strings.HasPrefix(long_url, "https://") {
		long_url = "https://" + long_url
	}
	http.Redirect(w, r, long_url, http.StatusFound)
}
