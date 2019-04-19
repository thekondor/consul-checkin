package consul_checkin

import (
	consul_api "github.com/hashicorp/consul/api"
)

// Service is a an entry-point of service to be checked-in to Consul.
type Service struct {
	ConsulService *consul_api.AgentServiceRegistration
}

// Config keeps a setup for Consul check-in.
type Config struct {
	Consul ConsulOptions
	On     CheckinEvents
}

// ConsulOptions keeps Consul API's genuine settings to interact with the agent.
type ConsulOptions struct {
	Client *consul_api.Client
}

// CheckinEvents holds callbacks used when a connection to Consul Agent is either established or lost.
type CheckinEvents struct {
	ServiceRegisterFailed   OnServiceRegisterFailedFunc
	ServiceDeregisterFailed OnServiceDeregisterFailedFunc
}

type (
	// OnServiceRegisterFailedFunc is an alias for a callback type used when a self-registration to Consul is failed.
	OnServiceRegisterFailedFunc func(serviceId string, err error)

	// OnServiceDeregisterFailedFunc is an aliase for a callback type used when a self-dregitration from Consule is failed.
	OnServiceDeregisterFailedFunc func(serviceId string, err error)
)
