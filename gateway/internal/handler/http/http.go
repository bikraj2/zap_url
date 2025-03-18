package httphandler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/bikraj2/url_shortener/gateway/internal"
	"github.com/bikraj2/url_shortener/gateway/internal/controller"
	"github.com/bikraj2/url_shortener/gateway/internal/repository/data"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type Handler struct {
	ctrl   *controller.Controller
	logger *zap.Logger
}

func New(ctrl *controller.Controller, logger *zap.Logger) *Handler {
	return &Handler{ctrl, logger}
}

func (h *Handler) CreateShortUrl(c *gin.Context) {
	var URL struct {
		LongUrl string `json:"long_url"`
	}

	if err := c.ShouldBindJSON(&URL); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON body"})
		return
	}

	shortURL, err := h.ctrl.CreateShortUrl(c.Request.Context(), URL.LongUrl)
	if err != nil {
		switch {
		case errors.Is(err, customerror.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Short URL not found"})
		case errors.Is(err, customerror.ErrBadRequest):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"short_url": shortURL})
}

func (h *Handler) Redirect(c *gin.Context) {
	c.Redirect(http.StatusTemporaryRedirect, "https://www.zapurl.tech/main")
}

func (h *Handler) GetLongUrl(c *gin.Context) {
	shortURL := c.Param("short_url")

	longURL, err := h.ctrl.GetLongUrl(c.Request.Context(), shortURL)
	if err != nil {
		customerror.RespondWithError(c, err)
		return
	}

	if !strings.HasPrefix(longURL, "http://") && !strings.HasPrefix(longURL, "https://") {
		longURL = "https://" + longURL
	}

	c.JSON(http.StatusOK, gin.H{"long_url": longURL})
}

func (h *Handler) RegisterUserHandler(c *gin.Context) {

	var user data.User

	if err := c.ShouldBindJSON(&user); err != nil {
		validationErrors := make(map[string]any)

		for _, fieldErr := range err.(validator.ValidationErrors) {
			validationErrors[fieldErr.Field()] = fmt.Sprintf("Field '%s' failed on '%s' rule", fieldErr.Field(), fieldErr.Tag())
		}
		customerror.RespondWithError(c, customerror.NewValidationError("Invalid input data", validationErrors))
		return
	}

	err := h.ctrl.RegisterUser(c.Request.Context(), &user)
	if err != nil {
		customerror.RespondWithError(c, err)
		return
	}
}
