package consul_checkin

import (
	consul_api "github.com/hashicorp/consul/api"
	. "github.com/thekondor/consul-checkin/watchdog"
	"time"
)

type AutoCheckinService struct {
	config             *Config
	consulAgent        *consul_api.Agent
	connectionWatchdog ConnectionWatchdog
	services           []Service
}

// TODO: check for already existing services with duplicate IDs
func (self *AutoCheckinService) Add(service Service) {
	self.services = append(self.services, service)
}

func (self *AutoCheckinService) Start() {
	self.connectionWatchdog = self.watchConsul()
	self.registerServices()
}

func (self *AutoCheckinService) Stop() {
	self.deregisterServices()
	self.connectionWatchdog.Stop()
}

func (self AutoCheckinService) watchConsul() ConnectionWatchdog {
	watchdog, _ := WatchConnection(
		CheckHealthOverLeaderPing(self.config.Consul.Client),
		&WatchOptions{
			PingInterval: 3 * time.Second,
			On: ConnectionEvents{
				ConnectionRecovered: self.registerServices,
				ConnectionFailed: func(uint, error) RetryDecision {
					return TryAgain
				},
				ConnectionLost: func(error) RetryDecision {
					return TryAgain
				},
			},
		},
	)
	return watchdog
}

func (self AutoCheckinService) registerServices() {
	for _, svc := range self.services {
		self.register(svc.ConsulService)
	}
}

func (self AutoCheckinService) deregisterServices() {
	for _, svc := range self.services {
		self.deregister(svc.ConsulService.ID)
	}
}

func (self AutoCheckinService) register(consulService *consul_api.AgentServiceRegistration) {
	err := self.consulAgent.ServiceRegister(consulService)
	if nil != err {
		self.config.On.ServiceRegisterFailed(consulService.ID, err)
	}
}

func (self AutoCheckinService) deregister(consulServiceId string) {
	err := self.consulAgent.ServiceDeregister(consulServiceId)
	if nil != err {
		self.config.On.ServiceDeregisterFailed(consulServiceId, err)
	}
}

func CheckinAutomatically(config *Config) *AutoCheckinService {
	return &AutoCheckinService{
		config:      config,
		consulAgent: config.Consul.Client.Agent(),
	}
}
