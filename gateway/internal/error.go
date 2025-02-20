package customerror

import (
	"errors"
	"net/http"

	"github.com/bikraj2/url_shortener/gateway/internal/helper"
)

var (
	ErrNotFound      = errors.New("the specified resource could not be found")
	ErrInternalError = errors.New("internal serve error")
	ErrBadRequest    = errors.New("request is badly formatted")
)

func ErrorResponse(w http.ResponseWriter, status int, headers http.Header, message string) {
	data := helper.Envelope{"error": message}
	err := helper.WriteJSON(w, status, data, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func NotFoundResponse(w http.ResponseWriter) {
	message := "thee resource you requested for doesnot exist"
	ErrorResponse(w, http.StatusNotFound, nil, message)
}
func BadRequestErrorResponse(w http.ResponseWriter) {
	ErrorResponse(w, http.StatusBadRequest, nil, ErrBadRequest.Error())
}
