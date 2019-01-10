package test_helpers

import (
	"github.com/stretchr/testify/mock"
)

type ConnectionHealthCheckMock struct {
	mock.Mock
}

func (mock *ConnectionHealthCheckMock) Ping() error {
	args := mock.Called()
	return args.Error(0)
}
