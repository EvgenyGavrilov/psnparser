// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package psnparser

import (
	models "github.com/EvgenyGavrilov/psnclient/models"
	mock "github.com/stretchr/testify/mock"
)

// MockPSNClienter is an autogenerated mock type for the PSNClienter type
type MockPSNClienter struct {
	mock.Mock
}

// ListGames provides a mock function with given fields: params
func (_m *MockPSNClienter) ListGames(params models.ListParams) (*models.ListGames, error) {
	ret := _m.Called(params)

	var r0 *models.ListGames
	if rf, ok := ret.Get(0).(func(models.ListParams) *models.ListGames); ok {
		r0 = rf(params)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.ListGames)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(models.ListParams) error); ok {
		r1 = rf(params)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ProductByURL provides a mock function with given fields: u
func (_m *MockPSNClienter) ProductByURL(u string) (*models.Product, error) {
	ret := _m.Called(u)

	var r0 *models.Product
	if rf, ok := ret.Get(0).(func(string) *models.Product); ok {
		r0 = rf(u)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Product)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(u)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
