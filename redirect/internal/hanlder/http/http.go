package http

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/bikraj2/url_shortener/redirect/internal/controller"
	"github.com/bikraj2/url_shortener/redirect/internal/repository/postgresql"
	"github.com/gorilla/mux"
)

type handler struct {
	ctrl *controller.Controller
}

type envelope map[string]interface{}

func New(ctrl *controller.Controller) *handler {
	return &handler{ctrl: ctrl}
}
func (h *handler) Handle(w http.ResponseWriter, r *http.Request) {
	short_url := mux.Vars(r)["short_url"]
	if short_url == "" {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Missing ID in URL"))
		return
	}
	long_url, err := h.ctrl.GetLongUrl(context.Background(), short_url)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		log.Print(err.Error())
		switch {
		case errors.Is(err, repository.ErrNotFound):
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("short url doesnot exist"))
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	err = writeJSON(w, http.StatusTemporaryRedirect, envelope{"long_url": long_url}, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
