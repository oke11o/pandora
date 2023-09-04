// Code generated by mockery v2.22.1. DO NOT EDIT.

package netsample

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
	core "github.com/yandex/pandora/core"
)

// MockAggregator is an autogenerated mock type for the Aggregator type
type MockAggregator struct {
	mock.Mock
}

// Report provides a mock function with given fields: sample
func (_m *MockAggregator) Report(sample *Sample) {
	_m.Called(sample)
}

// Run provides a mock function with given fields: ctx, deps
func (_m *MockAggregator) Run(ctx context.Context, deps core.AggregatorDeps) error {
	ret := _m.Called(ctx, deps)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, core.AggregatorDeps) error); ok {
		r0 = rf(ctx, deps)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewMockAggregator interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockAggregator creates a new instance of MockAggregator. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockAggregator(t mockConstructorTestingTNewMockAggregator) *MockAggregator {
	mock := &MockAggregator{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
