package discovery

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

type Registry interface {
	RegisterService(ctx context.Context, serviceName string, instanceID string, hostPort string) error
	DeRegisterService(ctx context.Context, serviceName string, instanceID string) error
	ServicesAddresses(ctx context.Context, serviceName string) ([]string, error)
	ReportHealthyState(serviceName string, instanceID string) error
}

var ErrNotFound = errors.New("no service address found")

func GenerateInstanceID(serviceName string) string {
	return fmt.Sprintf("%s-%d", serviceName, rand.New(rand.NewSource(time.Now().UnixNano())))
}
