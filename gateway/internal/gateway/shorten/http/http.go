package shorten

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"

	customerror "github.com/bikraj2/url_shortener/gateway/internal"
	"github.com/bikraj2/url_shortener/pkg/discovery"
)

type Gateway struct {
	registry discovery.Registry
}

const (
	ServiceName = "shorten"
)

func New(registry discovery.Registry) *Gateway {
	return &Gateway{registry: registry}
}

func (r *Gateway) CreateShortUrl(ctx context.Context, long_url string) (string, error) {

	addrs, err := r.registry.ServicesAddresses(ctx, ServiceName)
	if err != nil {
		return "", err
	}

	url := "http://" + addrs[rand.Intn(len(addrs))]
	body := struct {
		Long_url string `json:"long_url"`
	}{Long_url: long_url}
	jsonData, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("%v:error encoding json data:%w", customerror.ErrInternalError, err)
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusBadRequest {
		return "", customerror.ErrBadRequest
	} else if resp.StatusCode/100 != 2 {
		return "", fmt.Errorf("error generating the URL")
	}

	var short_url struct {
		Short_url string `json:"short_url"`
	}
	err = json.NewDecoder(resp.Body).Decode(&short_url)
	if err != nil {
		return "", fmt.Errorf("%v:%v", customerror.ErrInternalError, err)
	}
	return short_url.Short_url, nil
}
