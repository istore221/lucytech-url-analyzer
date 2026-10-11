package mocks

import (
	"context"

	"github.com/istore221/lucytech-url-analyzer/internal/analyzer"
	"github.com/stretchr/testify/mock"
)

type MockScraperEngine struct {
	mock.Mock
}

func (_m *MockScraperEngine) Analyze(ctx context.Context, targetURL string) (*analyzer.PageAnalysis, error) {
	ret := _m.Called(ctx, targetURL)

	if len(ret) == 0 {
		panic("no return value specified for Analyze")
	}

	var r0 *analyzer.PageAnalysis
	var r1 error

	if rf, ok := ret.Get(0).(func(context.Context, string) (*analyzer.PageAnalysis, error)); ok {
		return rf(ctx, targetURL)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *analyzer.PageAnalysis); ok {
		r0 = rf(ctx, targetURL)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*analyzer.PageAnalysis)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, targetURL)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

func NewMockScraperEngine(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockScraperEngine {
	mockEngine := &MockScraperEngine{}
	mockEngine.Mock.Test(t)

	t.Cleanup(func() { mockEngine.AssertExpectations(t) })

	return mockEngine
}
