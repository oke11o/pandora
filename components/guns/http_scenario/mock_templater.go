// Code generated by mockery v2.22.1. DO NOT EDIT.

package httpscenario

import mock "github.com/stretchr/testify/mock"

// MockTemplater is an autogenerated mock type for the Templater type
type MockTemplater struct {
	mock.Mock
}

// Apply provides a mock function with given fields: request, variables, scenarioName, stepName
func (_m *MockTemplater) Apply(request *RequestParts, variables map[string]interface{}, scenarioName string, stepName string) error {
	ret := _m.Called(request, variables, scenarioName, stepName)

	var r0 error
	if rf, ok := ret.Get(0).(func(*RequestParts, map[string]interface{}, string, string) error); ok {
		r0 = rf(request, variables, scenarioName, stepName)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewMockTemplater interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockTemplater creates a new instance of MockTemplater. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockTemplater(t mockConstructorTestingTNewMockTemplater) *MockTemplater {
	mock := &MockTemplater{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
