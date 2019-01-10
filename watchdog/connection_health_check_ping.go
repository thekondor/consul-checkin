package watchdog

import (
	consul_api "github.com/hashicorp/consul/api"
)

type healthCheckOverLeaderPing struct {
	consulClient *consul_api.Client
}

func (self healthCheckOverLeaderPing) Ping() error {
	_, err := self.consulClient.Status().Leader()
	return err
}

func CheckHealthOverLeaderPing(consulClient *consul_api.Client) ConnectionHealthCheck {
	return &healthCheckOverLeaderPing{consulClient}
}
