package consul

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	consul "github.com/hashicorp/consul/api"
)

type Registry struct {
	client *consul.Client
}

func New(addr string) (*Registry, error) {
	config := consul.DefaultConfig()
	config.Address = addr
	client, err := consul.NewClient(config)
	if err != nil {
		return nil, err
	}
	return &Registry{client: client}, nil
}

func (r *Registry) RegisterService(ctx context.Context, serviceName string, instanceID string, hostPort string) error {
	parts := strings.Split(hostPort, ":")
	if len(parts) != 2 {
		panic("host port must be of the format : <host>:<port> exmaple: localhost:4000")
	}
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return err
	}

	return r.client.Agent().ServiceRegister(&consul.AgentServiceRegistration{
		Address: parts[0],
		Name:    serviceName,
		ID:      instanceID,
		Port:    port,
		Check:   &consul.AgentServiceCheck{CheckID: instanceID, TTL: "5s"},
	})
}

func (r *Registry) DeRegisterService(ctx context.Context, _ string, instanceID string) error {
	return r.client.Agent().ServiceDeregister(instanceID)
}
func (r *Registry) ServicesAddresses(ctx context.Context, serviceName string) ([]string, error) {
	entries, _, err := r.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, err
	}

	var res []string
	for _, e := range entries {
		res = append(res, fmt.Sprintf("%v:%v", e.Service.Address, e.Service.Port))
	}
	return res, nil
}

func (r *Registry) ReportHealthyState(instanceID string, _ string) error {
	return r.client.Agent().UpdateTTL(instanceID, "", "pass")
}
