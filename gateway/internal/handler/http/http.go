package httphandler

import (
	"errors"
	"net/http"
	"strings"

	customerror "github.com/bikraj2/url_shortener/gateway/internal"
	"github.com/bikraj2/url_shortener/gateway/internal/controller"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	ctrl *controller.Controller
}

func New(ctrl *controller.Controller) *Handler {
	return &Handler{ctrl}
}

// ✅ Convert `CreateShortUrl` to use Gin
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
		switch {
		case errors.Is(err, customerror.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Short URL not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	if !strings.HasPrefix(longURL, "http://") && !strings.HasPrefix(longURL, "https://") {
		longURL = "https://" + longURL
	}

	c.JSON(http.StatusOK, gin.H{"long_url": longURL})
}
