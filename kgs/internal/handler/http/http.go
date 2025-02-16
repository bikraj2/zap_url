package http

import (
	"context"
	"log"
	"net/http"

	"github.com/bikraj2/url_shortener/kgs/internal/controller"
)

type Handler struct {
	ctrl *controller.Controller
}

type envelope map[string]interface{}

func New(ctrl *controller.Controller) *Handler {
	return &Handler{ctrl}
}

func (h *Handler) Handle(w http.ResponseWriter, _ *http.Request) {
	key, err := h.ctrl.GetNewKey(context.Background())

	if err != nil {
		log.Println(err.Error())
		err = writeJSON(w, http.StatusInternalServerError, err.Error(), nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return

	}
	err = writeJSON(w, http.StatusOK, envelope{"key": key}, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
