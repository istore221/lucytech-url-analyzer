package mocks

import (
	"context"

	"github.com/istore221/lucytech-url-analyzer/cmd/http/internal/models"
	"github.com/stretchr/testify/mock"
)

type MockEngine struct {
	mock.Mock
}

func (_m *MockEngine) Analyze(ctx context.Context, targetURL string) (*models.PageAnalysis, error) {
	ret := _m.Called(ctx, targetURL)

	if len(ret) == 0 {
		panic("no return value specified for Analyze")
	}

	var r0 *models.PageAnalysis
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string) (*models.PageAnalysis, error)); ok {
		return rf(ctx, targetURL)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *models.PageAnalysis); ok {
		r0 = rf(ctx, targetURL)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.PageAnalysis)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, targetURL)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

func NewMockEngine(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockEngine {
	mockEngine := &MockEngine{}
	mockEngine.Mock.Test(t)

	t.Cleanup(func() { mockEngine.AssertExpectations(t) })

	return mockEngine
}
