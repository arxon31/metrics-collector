// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package v3

import (
	"sync"
)

// Ensure, that pingerServiceMock does implement pingerService.
// If this is not the case, regenerate this file with moq.
var _ pingerService = &pingerServiceMock{}

// pingerServiceMock is a mock implementation of pingerService.
//
//	func TestSomethingThatUsespingerService(t *testing.T) {
//
//		// make and configure a mocked pingerService
//		mockedpingerService := &pingerServiceMock{
//			PingDBFunc: func() error {
//				panic("mock out the PingDB method")
//			},
//		}
//
//		// use mockedpingerService in code that requires pingerService
//		// and then make assertions.
//
//	}
type pingerServiceMock struct {
	// PingDBFunc mocks the PingDB method.
	PingDBFunc func() error

	// calls tracks calls to the methods.
	calls struct {
		// PingDB holds details about calls to the PingDB method.
		PingDB []struct {
		}
	}
	lockPingDB sync.RWMutex
}

// PingDB calls PingDBFunc.
func (mock *pingerServiceMock) PingDB() error {
	if mock.PingDBFunc == nil {
		panic("pingerServiceMock.PingDBFunc: method is nil but pingerService.PingDB was just called")
	}
	callInfo := struct {
	}{}
	mock.lockPingDB.Lock()
	mock.calls.PingDB = append(mock.calls.PingDB, callInfo)
	mock.lockPingDB.Unlock()
	return mock.PingDBFunc()
}

// PingDBCalls gets all the calls that were made to PingDB.
// Check the length with:
//
//	len(mockedpingerService.PingDBCalls())
func (mock *pingerServiceMock) PingDBCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockPingDB.RLock()
	calls = mock.calls.PingDB
	mock.lockPingDB.RUnlock()
	return calls
}