package redirect

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/bikraj2/url_shortener/pkg/discovery"
	customerror "github.com/bikraj2/url_shortener/shortener-gateway/internal"
)

type Gateway struct {
	registry discovery.Registry
}

const (
	ServiceName = "redirect"
)

func New(registry discovery.Registry) *Gateway {
	return &Gateway{registry: registry}
}

func (r *Gateway) GetLongUrl(ctx context.Context, short_url string) (string, error) {

	addrs, err := r.registry.ServicesAddresses(ctx, ServiceName)
	if err != nil {
		return "", err
	}

	url := "http://" + addrs[rand.Intn(len(addrs))] + "/" + short_url
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req = req.WithContext(ctx)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return "", customerror.ErrNotFound
	} else if resp.StatusCode/100 == 3 {
		var URL struct {
			LongUrl string `json:"long_url"`
		}
		err := json.NewDecoder(resp.Body).Decode(&URL)
		if err != nil {
			return "", err
		}
		return URL.LongUrl, nil
	} else {
		return "", fmt.Errorf("non-2xx respons: %v", resp)
	}
}
