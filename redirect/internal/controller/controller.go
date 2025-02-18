package controller

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	// "github.com/bikraj2/url_shortener/redirect/internal/repository/postgresql"
	"math/rand"
	"net"
	"syscall"
	"time"
)

type redirectRepository interface {
	GetLongUrl(context.Context, string) (string, error)
}

type Controller struct {
	repo redirectRepository
}

func New(repo redirectRepository) *Controller {
	return &Controller{repo: repo}
}

// GetLongUrl with automatic retries on transient errors
func (ctrl *Controller) GetLongUrl(ctx context.Context, short_url string) (string, error) {
	const maxRetries = 5
	baseDelay := time.Millisecond * 10

	var long_url string
	var err error

	for attempt := 0; attempt < maxRetries; attempt++ {
		long_url, err = ctrl.repo.GetLongUrl(ctx, short_url)
		if err == nil {
			return long_url, nil
		}

		if !shouldRetry(err) {
			return "", err
		}

		backoff := baseDelay * time.Duration(1<<attempt)
		jitter := time.Duration(rand.Intn(100)) * time.Millisecond
		sleepDuration := backoff + jitter

		fmt.Printf("Retrying due to error: %v (attempt %d/%d, waiting %v)\n", err, attempt+1, maxRetries, sleepDuration)
		time.Sleep(sleepDuration)
	}

	return "", fmt.Errorf("failed to get long URL after %d attempts: %w", maxRetries, err)
}

func shouldRetry(err error) bool {
	if err == nil {
		return false
	}

	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return true
	case errors.Is(err, context.Canceled):
		return false

	case errors.Is(err, sql.ErrConnDone):
		return true
	case errors.Is(err, sql.ErrTxDone):
		return false

	case isNetworkError(err):
		return true

	default:
		return false
	}
}

func isNetworkError(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}

	var syscallErr syscall.Errno
	if errors.As(err, &syscallErr) {
		return syscallErr == syscall.ECONNRESET || syscallErr == syscall.ECONNABORTED || syscallErr == syscall.ETIMEDOUT
	}

	return false
}
