// Code generated by mockery v2.39.1. DO NOT EDIT.

package mockery

import mock "github.com/stretchr/testify/mock"

// MockConfiguration_configuration is an autogenerated mock type for the Configuration type
type MockConfiguration_configuration struct {
	mock.Mock
}

type MockConfiguration_configuration_Expecter struct {
	mock *mock.Mock
}

func (_m *MockConfiguration_configuration) EXPECT() *MockConfiguration_configuration_Expecter {
	return &MockConfiguration_configuration_Expecter{mock: &_m.Mock}
}

// IndentSize provides a mock function with given fields:
func (_m *MockConfiguration_configuration) IndentSize() int {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for IndentSize")
	}

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// MockConfiguration_configuration_IndentSize_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IndentSize'
type MockConfiguration_configuration_IndentSize_Call struct {
	*mock.Call
}

// IndentSize is a helper method to define mock.On call
func (_e *MockConfiguration_configuration_Expecter) IndentSize() *MockConfiguration_configuration_IndentSize_Call {
	return &MockConfiguration_configuration_IndentSize_Call{Call: _e.mock.On("IndentSize")}
}

func (_c *MockConfiguration_configuration_IndentSize_Call) Run(run func()) *MockConfiguration_configuration_IndentSize_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockConfiguration_configuration_IndentSize_Call) Return(_a0 int) *MockConfiguration_configuration_IndentSize_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockConfiguration_configuration_IndentSize_Call) RunAndReturn(run func() int) *MockConfiguration_configuration_IndentSize_Call {
	_c.Call.Return(run)
	return _c
}

// OneBracketPerLine provides a mock function with given fields:
func (_m *MockConfiguration_configuration) OneBracketPerLine() bool {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for OneBracketPerLine")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// MockConfiguration_configuration_OneBracketPerLine_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'OneBracketPerLine'
type MockConfiguration_configuration_OneBracketPerLine_Call struct {
	*mock.Call
}

// OneBracketPerLine is a helper method to define mock.On call
func (_e *MockConfiguration_configuration_Expecter) OneBracketPerLine() *MockConfiguration_configuration_OneBracketPerLine_Call {
	return &MockConfiguration_configuration_OneBracketPerLine_Call{Call: _e.mock.On("OneBracketPerLine")}
}

func (_c *MockConfiguration_configuration_OneBracketPerLine_Call) Run(run func()) *MockConfiguration_configuration_OneBracketPerLine_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockConfiguration_configuration_OneBracketPerLine_Call) Return(_a0 bool) *MockConfiguration_configuration_OneBracketPerLine_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockConfiguration_configuration_OneBracketPerLine_Call) RunAndReturn(run func() bool) *MockConfiguration_configuration_OneBracketPerLine_Call {
	_c.Call.Return(run)
	return _c
}

// TrimMultipleEmptyLines provides a mock function with given fields:
func (_m *MockConfiguration_configuration) TrimMultipleEmptyLines() bool {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for TrimMultipleEmptyLines")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// MockConfiguration_configuration_TrimMultipleEmptyLines_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'TrimMultipleEmptyLines'
type MockConfiguration_configuration_TrimMultipleEmptyLines_Call struct {
	*mock.Call
}

// TrimMultipleEmptyLines is a helper method to define mock.On call
func (_e *MockConfiguration_configuration_Expecter) TrimMultipleEmptyLines() *MockConfiguration_configuration_TrimMultipleEmptyLines_Call {
	return &MockConfiguration_configuration_TrimMultipleEmptyLines_Call{Call: _e.mock.On("TrimMultipleEmptyLines")}
}

func (_c *MockConfiguration_configuration_TrimMultipleEmptyLines_Call) Run(run func()) *MockConfiguration_configuration_TrimMultipleEmptyLines_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockConfiguration_configuration_TrimMultipleEmptyLines_Call) Return(_a0 bool) *MockConfiguration_configuration_TrimMultipleEmptyLines_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockConfiguration_configuration_TrimMultipleEmptyLines_Call) RunAndReturn(run func() bool) *MockConfiguration_configuration_TrimMultipleEmptyLines_Call {
	_c.Call.Return(run)
	return _c
}

// UseTabs provides a mock function with given fields:
func (_m *MockConfiguration_configuration) UseTabs() bool {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for UseTabs")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// MockConfiguration_configuration_UseTabs_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UseTabs'
type MockConfiguration_configuration_UseTabs_Call struct {
	*mock.Call
}

// UseTabs is a helper method to define mock.On call
func (_e *MockConfiguration_configuration_Expecter) UseTabs() *MockConfiguration_configuration_UseTabs_Call {
	return &MockConfiguration_configuration_UseTabs_Call{Call: _e.mock.On("UseTabs")}
}

func (_c *MockConfiguration_configuration_UseTabs_Call) Run(run func()) *MockConfiguration_configuration_UseTabs_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockConfiguration_configuration_UseTabs_Call) Return(_a0 bool) *MockConfiguration_configuration_UseTabs_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockConfiguration_configuration_UseTabs_Call) RunAndReturn(run func() bool) *MockConfiguration_configuration_UseTabs_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockConfiguration_configuration creates a new instance of MockConfiguration_configuration. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockConfiguration_configuration(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockConfiguration_configuration {
	mock := &MockConfiguration_configuration{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
