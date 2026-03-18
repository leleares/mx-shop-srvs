package consul

import (
	"fmt"

	"github.com/hashicorp/consul/api"
)

type Registry struct {
	Host string
	Port int
}

type RegistryClient interface {
	Register(address string, port int, id string, name string, tags []string) error
	DeRegister(serviceID string) error
}

func NewRegistryClient(host string, port int) RegistryClient {
	return &Registry{
		Host: host,
		Port: port,
	}
}

func (r *Registry) Register(address string, port int, id string, name string, tags []string) error {
	conf := api.DefaultConfig()
	conf.Address = fmt.Sprintf("%s:%d", r.Host, r.Port)

	client, err := api.NewClient(conf)
	if err != nil {
		panic(err)
	}

	check := api.AgentServiceCheck{
		GRPC:                           fmt.Sprintf("%s:%d", address, port),
		Timeout:                        "5s",
		Interval:                       "5s",
		DeregisterCriticalServiceAfter: "10s",
	}

	registration := new(api.AgentServiceRegistration)
	registration.ID = id
	registration.Address = address
	registration.Name = name
	registration.Port = port
	registration.Tags = tags
	registration.Check = &check

	err = client.Agent().ServiceRegister(registration)
	if err != nil {
		panic(err)
	}
	return nil
}

func (r *Registry) DeRegister(serviceId string) error {
	conf := api.DefaultConfig()
	conf.Address = fmt.Sprintf("%s:%d", r.Host, r.Port)

	client, err := api.NewClient(conf)
	if err != nil {
		return err
	}
	return client.Agent().ServiceDeregister(serviceId)
}
