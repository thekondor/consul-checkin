# consul-checkin

> This is a dummy draft implementation of Go/Golang library to register Golang services automatically within [Hashicorp Consul](https://www.consul.io/). This one works but definitely expected to be unstable as well as provides unclear API which is a subject to change. The library is made public available since it became dependency for already deployed production services.

## Rationale

HashiCorp [Consul's API](https://godoc.org/github.com/hashicorp/consul/api) allows to register service in a simple manner. Due to the nature the API is limited:

1. as soon as service is registered on Consul's agent, nothing happens when Consul Agent become unavailable for some time due to any reason; i.e. the service wouldn't be re-registered again (especially when Consul cluster goes down).

2. when a Golang service starts before Consul Agent is available, it has to watch its availability to register itself using API above provided.

`consul-checkin` is a Go/Golang library to watch connection to a Consul agent and register service as soon as Consul agent becomes available; or re-register when a connection to a Consul agent is lost. **Such strategy allows microservices to be deployed independently of Consul Agent's state**.

## Usage

```golang
import consul_checkin "github.com/thekondor/consul-checkin"
import consul_api "github.com/hashicorp/consul/api"

func main() {
...

// create consul client using default's HashiCorp API
consulClient, err := consul_api.NewClient(consul_api.DefaultConfig())
if nil != err {
  log.Fatalf("Failed to create consul client, err = %+v", err)
}

// create a worker to checkin service automatically as soon as connection with Consul is established or recovered 
selfCheckin := consul_checkin.CheckinAutomatically(
  &consul_checkin.Config{
    Consul: consul_checkin.ConsulOptions{consulclient},
    On: CheckinEvents{
      ServiceRegisterFailed: func(serviceId string, err error) {
        log.Printf("Service registration '%s' failed due to '%v'", serviceId, err)
      },
      ServiceDeregisterFailed: func(serviceId string, err error) {
        log.Printf("Service de-registration '%s' failed due to '%v'", serviceId, err)
      },
    },
  },
)

// Declare services to register
selfCheckin.Add(consul_checkin.Service{
  ConsulService: &consul_api.AgentServiceRegistration{
   ... // to be initialized accordingly to github.com/hashicorp/consul/api
  },
})

// Start services registration in a background. The call returns immediatelly.
selfCheckin.Start()

// In case of graceful shutdown, try to unregister services. Otherwise Consul would mark this application as a one with a critical state.
go func() {
  <- ctx.Done() // or other way to get notification when the application stops
  selfCheckin.Stop() // to unregister added service(s) on application stop
}()
...
}
```

# License

The library is released under the MIT license. See LICENSE file.
