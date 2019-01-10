package consul_checkin

import (
	consul_api "github.com/hashicorp/consul/api"
)

type Service struct {
	ConsulService *consul_api.AgentServiceRegistration
}

type Config struct {
	Consul ConsulOptions
	On     CheckinEvents
}

type ConsulOptions struct {
	Client *consul_api.Client
}

type CheckinEvents struct {
	ServiceRegisterFailed   OnServiceRegisterFailedFunc
	ServiceDeregisterFailed OnServiceDeregisterFailedFunc
}

type OnServiceRegisterFailedFunc func(serviceId string, err error)
type OnServiceDeregisterFailedFunc func(serviceId string, err error)
