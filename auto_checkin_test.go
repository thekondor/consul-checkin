package consul_checkin

import (
	consul_api "github.com/hashicorp/consul/api"
	"github.com/ory/dockertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

type AutoCheckinTest struct {
	suite.Suite

	consulClient *consul_api.Client

	docker struct {
		pool           *dockertest.Pool
		consulInstance *dockertest.Resource
		consulAddr     string
	}
	assert  *assert.Assertions
	require *require.Assertions
}

func TestAutoCheckin(t *testing.T) {
	suite.Run(t, new(AutoCheckinTest))
}

func (test *AutoCheckinTest) SetupTest() {
	test.assert, test.require = test.Assert(), test.Require()

	var err error
	test.docker.pool, err = dockertest.NewPool("")
	test.require.NoError(err)

	test.docker.consulInstance, test.docker.consulAddr = test.runConsulContainer()
	test.consulClient = test.prepareConsulClient()
	test.waitConsulContainerReady()
}

func (test *AutoCheckinTest) runConsulContainer() (*dockertest.Resource, string) {
	dockerOptions := dockertest.RunOptions{
		Repository: "consul",
		Tag:        "latest",
		Cmd:        []string{"agent", "-dev", "-client", "0.0.0.0"},
	}
	consulInstance, err := test.docker.pool.RunWithOptions(&dockerOptions)
	test.require.NoError(err)

	consulAddr := consulInstance.GetHostPort("8500/tcp")
	test.T().Logf("Consul instance available at %s", consulAddr)
	return consulInstance, consulAddr
}

func (test *AutoCheckinTest) waitConsulContainerReady() {
	err := test.docker.pool.Retry(func() error {
		_, err := test.consulClient.Agent().Services()
		return err
	})
	test.require.NoError(err, "Consul container is not responding")
}

func (test *AutoCheckinTest) prepareConsulClient() *consul_api.Client {
	consulClient, err := consul_api.NewClient(
		&consul_api.Config{
			Address: test.docker.consulAddr,
		},
	)
	test.require.NoError(err)

	return consulClient
}

func (test *AutoCheckinTest) TearDownTest() {
	err := test.docker.pool.Purge(test.docker.consulInstance)
	test.Require().NoError(err)
}

// TODO: this test tests too much cases. Should be refactored.
func (test *AutoCheckinTest) Test_NaiveAndQuick_AutoCheckin_InSimplestWayEver() {
	currentServices, err := test.consulClient.Agent().Services()
	test.require.NoError(err)
	test.require.Empty(currentServices, "No other services is supposed to be registered on Consul instance")

	sut := CheckinAutomatically(
		&Config{
			Consul: ConsulOptions{
				Client: test.consulClient,
			},
			On: CheckinEvents{
				ServiceRegisterFailed: func(serviceId string, err error) {
					test.require.Fail("service should be registered")
				},
				ServiceDeregisterFailed: func(serviceId string, err error) {
					test.require.Fail("service should be unregistered")
				},
			},
		},
	)

	sut.Add(Service{
		ConsulService: &consul_api.AgentServiceRegistration{
			ID:      "test-service-id",
			Name:    "service name",
			Port:    9090,
			Address: "localhost",
			Check: &consul_api.AgentServiceCheck{
				HTTP:                           "http://127.0.0.1:9090/checkme",
				Interval:                       "5s",
				Timeout:                        "3s",
				DeregisterCriticalServiceAfter: "1m",
			},
		},
	})
	sut.Start()

	currentServices, err = test.consulClient.Agent().Services()
	test.require.NoError(err)

	test.assert.NotEmpty(currentServices)
	test.assert.Contains(currentServices, "test-service-id")

	sut.Stop()

	currentServices, err = test.consulClient.Agent().Services()
	test.require.NoError(err)
	test.assert.Emptyf(currentServices, "the service is expected deregistered at stop")
}
