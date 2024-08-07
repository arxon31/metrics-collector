// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package v1

import (
	"context"
	"sync"
)

// Ensure, that providerServiceMock does implement providerService.
// If this is not the case, regenerate this file with moq.
var _ providerService = &providerServiceMock{}

// providerServiceMock is a mock implementation of providerService.
//
//	func TestSomethingThatUsesproviderService(t *testing.T) {
//
//		// make and configure a mocked providerService
//		mockedproviderService := &providerServiceMock{
//			GetCounterValueFunc: func(ctx context.Context, name string) (int64, error) {
//				panic("mock out the GetCounterValue method")
//			},
//			GetGaugeValueFunc: func(ctx context.Context, name string) (float64, error) {
//				panic("mock out the GetGaugeValue method")
//			},
//		}
//
//		// use mockedproviderService in code that requires providerService
//		// and then make assertions.
//
//	}
type providerServiceMock struct {
	// GetCounterValueFunc mocks the GetCounterValue method.
	GetCounterValueFunc func(ctx context.Context, name string) (int64, error)

	// GetGaugeValueFunc mocks the GetGaugeValue method.
	GetGaugeValueFunc func(ctx context.Context, name string) (float64, error)

	// calls tracks calls to the methods.
	calls struct {
		// GetCounterValue holds details about calls to the GetCounterValue method.
		GetCounterValue []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Name is the name argument value.
			Name string
		}
		// GetGaugeValue holds details about calls to the GetGaugeValue method.
		GetGaugeValue []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Name is the name argument value.
			Name string
		}
	}
	lockGetCounterValue sync.RWMutex
	lockGetGaugeValue   sync.RWMutex
}

// GetCounterValue calls GetCounterValueFunc.
func (mock *providerServiceMock) GetCounterValue(ctx context.Context, name string) (int64, error) {
	if mock.GetCounterValueFunc == nil {
		panic("providerServiceMock.GetCounterValueFunc: method is nil but providerService.GetCounterValue was just called")
	}
	callInfo := struct {
		Ctx  context.Context
		Name string
	}{
		Ctx:  ctx,
		Name: name,
	}
	mock.lockGetCounterValue.Lock()
	mock.calls.GetCounterValue = append(mock.calls.GetCounterValue, callInfo)
	mock.lockGetCounterValue.Unlock()
	return mock.GetCounterValueFunc(ctx, name)
}

// GetCounterValueCalls gets all the calls that were made to GetCounterValue.
// Check the length with:
//
//	len(mockedproviderService.GetCounterValueCalls())
func (mock *providerServiceMock) GetCounterValueCalls() []struct {
	Ctx  context.Context
	Name string
} {
	var calls []struct {
		Ctx  context.Context
		Name string
	}
	mock.lockGetCounterValue.RLock()
	calls = mock.calls.GetCounterValue
	mock.lockGetCounterValue.RUnlock()
	return calls
}

// GetGaugeValue calls GetGaugeValueFunc.
func (mock *providerServiceMock) GetGaugeValue(ctx context.Context, name string) (float64, error) {
	if mock.GetGaugeValueFunc == nil {
		panic("providerServiceMock.GetGaugeValueFunc: method is nil but providerService.GetGaugeValue was just called")
	}
	callInfo := struct {
		Ctx  context.Context
		Name string
	}{
		Ctx:  ctx,
		Name: name,
	}
	mock.lockGetGaugeValue.Lock()
	mock.calls.GetGaugeValue = append(mock.calls.GetGaugeValue, callInfo)
	mock.lockGetGaugeValue.Unlock()
	return mock.GetGaugeValueFunc(ctx, name)
}

// GetGaugeValueCalls gets all the calls that were made to GetGaugeValue.
// Check the length with:
//
//	len(mockedproviderService.GetGaugeValueCalls())
func (mock *providerServiceMock) GetGaugeValueCalls() []struct {
	Ctx  context.Context
	Name string
} {
	var calls []struct {
		Ctx  context.Context
		Name string
	}
	mock.lockGetGaugeValue.RLock()
	calls = mock.calls.GetGaugeValue
	mock.lockGetGaugeValue.RUnlock()
	return calls
}
