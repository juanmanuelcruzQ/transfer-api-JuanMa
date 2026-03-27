package mocks

import mock "github.com/stretchr/testify/mock"

type TransfersPublisherMock struct {
	mock.Mock
}

type TransfersPublisherMock_Expecter struct {
	mock *mock.Mock
}

func (_m *TransfersPublisherMock) EXPECT() *TransfersPublisherMock_Expecter {
	return &TransfersPublisherMock_Expecter{mock: &_m.Mock}
}

func (_m *TransfersPublisherMock) Publish(operation string, transferID string) error {
	args := _m.Called(operation, transferID)

	if len(args) == 0 {
		panic("no return value specified for Publish")
	}

	var r0 error
	if rf, ok := args.Get(0).(func(string, string) error); ok {
		r0 = rf(operation, transferID)
	} else {
		r0 = args.Error(0)
	}

	return r0
}

type TransfersPublisherMock_Publish_Call struct {
	*mock.Call
}

func (_e *TransfersPublisherMock_Expecter) Publish(operation interface{}, transferID interface{}) *TransfersPublisherMock_Publish_Call {
	return &TransfersPublisherMock_Publish_Call{Call: _e.mock.On("Publish", operation, transferID)}
}

func (_c *TransfersPublisherMock_Publish_Call) Run(run func(operation string, transferID string)) *TransfersPublisherMock_Publish_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string))
	})
	return _c
}

func (_c *TransfersPublisherMock_Publish_Call) Return(_a0 error) *TransfersPublisherMock_Publish_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *TransfersPublisherMock_Publish_Call) RunAndReturn(run func(string, string) error) *TransfersPublisherMock_Publish_Call {
	_c.Call.Return(run)
	return _c
}

func NewTransfersPublisherMock(t interface {
	mock.TestingT
	Cleanup(func())
}) *TransfersPublisherMock {
	mock := &TransfersPublisherMock{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
