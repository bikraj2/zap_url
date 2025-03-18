package customerror

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	ErrNotFound      = NewNotFoundError("The specified resource could not be found", nil)
	ErrInternalError = NewInternalServerError("Internal server error", nil)
	ErrBadRequest    = NewValidationError("Request is badly formatted", nil)
)

type AppError struct {
	HTTPCode int
	Code     string
	Message  string
	Details  map[string]any
}

func (err *AppError) Error() string {
	return fmt.Sprintf("Code: %s | HTTPCode: %d | Message: %s | Details: %v", err.Code, err.HTTPCode, err.Message, err.Details)
}

func RespondWithError(c *gin.Context, err error) {
	if appErr, ok := err.(*AppError); ok {
		c.JSON(appErr.HTTPCode, gin.H{
			"error":   appErr.Code,
			"message": appErr.Message,
			"details": appErr.Details,
		})
	} else {
		// If it's a generic Go error, wrap it into an internal server error
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "INTERNAL_SERVER_ERROR",
			"message": err.Error(),
		})
	}
}

func NewValidationError(msg string, details map[string]any) *AppError {
	return &AppError{HTTPCode: http.StatusBadRequest, Code: "VALIDATION_ERROR", Message: msg, Details: details}
}

func NewDuplicateError(msg string, details map[string]any) *AppError {
	return &AppError{HTTPCode: http.StatusConflict, Code: "DUPLICATE_ENTRY", Message: msg, Details: details}
}

func NewNotFoundError(msg string, details map[string]any) *AppError {
	return &AppError{HTTPCode: http.StatusNotFound, Code: "NOT_FOUND", Message: msg, Details: details}
}

func NewDatabaseError(msg string, details map[string]any) *AppError {
	return &AppError{HTTPCode: http.StatusInternalServerError, Code: "DB_ERROR", Message: msg, Details: details}
}

func NewUnauthorizedError(msg string, details map[string]any) *AppError {
	return &AppError{HTTPCode: http.StatusUnauthorized, Code: "UNAUTHORIZED", Message: msg, Details: details}
}

func NewInternalServerError(msg string, details map[string]any) *AppError {
	return &AppError{HTTPCode: http.StatusInternalServerError, Code: "INTERNAL_SERVER_ERROR", Message: msg, Details: details}
}
