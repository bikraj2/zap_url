package controller

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/bikraj2/url_shortener/kgs/internal/repository"
)

type kgsRepository interface {
	GetNewKey(ctx context.Context) (string, error)
}
type Controller struct {
	repo kgsRepository
}

const maxRetries = 5

func New(repo kgsRepository) *Controller {
	return &Controller{repo}
}

func (ctrl *Controller) GetNewKey(ctx context.Context) (string, error) {
	key, err := ctrl.repo.GetNewKey(ctx)
	for attempt := 1; attempt <= maxRetries; attempt++ {
		key, err = ctrl.repo.GetNewKey(ctx)
		if err == nil {
			return key, nil // Success, return key
		}

		// If no keys are left, retry after some delay
		if errors.Is(err, repository.ErrNoKeyleft) {
			if attempt < maxRetries {
				fmt.Printf("No key left, retrying... (attempt %d/%d)\n", attempt, maxRetries)
				random_delay_time := rand.Intn(10000) % 1000
				time.Sleep(time.Duration(random_delay_time * int(time.Millisecond))) // Wait before retrying
				continue
			}
			return "", fmt.Errorf("no key left after %d retries, please try again later", maxRetries)
		}
		return "", err
	}

	return key, nil
}
