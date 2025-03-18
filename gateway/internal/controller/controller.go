package controller

import (
	"context"

	redirect "github.com/bikraj2/url_shortener/gateway/internal/gateway/redirect/http"
	shorten "github.com/bikraj2/url_shortener/gateway/internal/gateway/shorten/http"
	"github.com/bikraj2/url_shortener/gateway/internal/repository/data"
)

type Controller struct {
	redirectGateway *redirect.Gateway
	shortenGateway  *shorten.Gateway
	userModel       *data.UserRepository
}

func New(rg *redirect.Gateway, sg *shorten.Gateway, userModel *data.UserRepository) *Controller {
	return &Controller{redirectGateway: rg, shortenGateway: sg, userModel: userModel}
}

func (ctrl *Controller) CreateShortUrl(ctx context.Context, long_url string) (string, error) {
	short_url, err := ctrl.shortenGateway.CreateShortUrl(ctx, long_url)
	if err != nil {
		return "", err
	}
	return short_url, nil
}

func (ctrl *Controller) GetLongUrl(ctx context.Context, short_url string) (string, error) {
	long_url, err := ctrl.redirectGateway.GetLongUrl(ctx, short_url)
	if err != nil {
		return "", err
	}
	return long_url, nil
}
func (ctrl *Controller) RegisterUser(ctx context.Context, user *data.User) error {
	return ctrl.userModel.RegisterUser(user)
}
