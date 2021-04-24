package psnparser

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/EvgenyGavrilov/psnclient/models"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestParser_Catalog(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		opt := Options{
			CountThreads:  2,
			CountElements: 100,
		}

		expectedList := map[string]*models.ListGames{
			"1": {
				ID: "1",
				Links: []models.ListGamesLink{
					{Bucket: "test1"},
					{Bucket: "test2"},
				},
			},
			"2": {
				ID: "2",
				Links: []models.ListGamesLink{
					{Bucket: "test3"},
					{Bucket: "test4"},
				},
			},
		}

		params := models.ListParams{
			Size:  100,
			Start: 0,
		}

		cli := &MockPSNClienter{}
		cli.On("ListGames", params).
			Return(expectedList["1"], nil).
			Once()

		params.Start = 100
		cli.On("ListGames", params).
			Return(expectedList["2"], nil).
			Once()

		cli.On("ListGames", mock.Anything).
			Return(&models.ListGames{}, nil).
			Twice()

		p := New(opt, cli)
		ch := p.Catalog(context.Background())

		done := make(chan bool)
		actual := make(map[string]*models.ListGames)

		go func(done chan<- bool, ch <-chan ResponseCatalog, actual map[string]*models.ListGames, t *testing.T) {
			timer := time.After(time.Second)
			for {
				select {
				case <-timer:
					done <- false
				case el, ok := <-ch:
					if !ok {
						done <- true
						return
					}
					require.NoError(t, el.Error)
					actual[el.Data.ID] = el.Data
				}
			}
		}(done, ch, actual, t)

		require.Truef(t, <-done, "Channel is not closed")
		require.Equal(t, expectedList, actual)

		cli.AssertExpectations(t)
	})

	t.Run("Context canceled", func(t *testing.T) {
		expectedList := &models.ListGames{
			ID: "1",
			Links: []models.ListGamesLink{
				{Bucket: "test"},
				{Bucket: "test"},
			},
		}

		cli := &MockPSNClienter{}
		cli.On("ListGames", mock.Anything).
			Return(expectedList, nil).
			Maybe()

		opt := Options{
			CountThreads:  2,
			CountElements: 100,
		}
		p := New(opt, cli)
		ctx, cancel := context.WithCancel(context.Background())
		ch := p.Catalog(ctx)

		done := make(chan bool)
		var actualErr error

		go func(done chan<- bool, ch <-chan ResponseCatalog, cancel context.CancelFunc) {
			timer := time.After(time.Second)
			for {
				select {
				case <-timer:
					done <- false
					return
				case el, ok := <-ch:
					if !ok {
						done <- true
						return
					}
					actualErr = el.Error
					cancel()
				}
			}
		}(done, ch, cancel)

		require.Truef(t, <-done, "Channel is not closed")
		require.ErrorIs(t, actualErr, context.Canceled)

		cli.AssertExpectations(t)
	})

	t.Run("Client return error", func(t *testing.T) {
		cli := &MockPSNClienter{}
		cli.On("ListGames", mock.Anything).
			Return(nil, errors.New("some error")).
			Once()

		opt := Options{
			CountThreads:  1,
			CountElements: 100,
		}
		p := New(opt, cli)
		ch := p.Catalog(context.Background())

		done := make(chan bool)

		go func(done chan<- bool, ch <-chan ResponseCatalog, t *testing.T) {
			timer := time.After(time.Second)
		loop:
			for {
				select {
				case <-timer:
					done <- false
				case el, ok := <-ch:
					if !ok {
						done <- true
						break loop
					}
					require.Error(t, el.Error)
				}
			}
		}(done, ch, t)

		require.Truef(t, <-done, "Channel is not closed")

		cli.AssertExpectations(t)
	})
}

func TestParser_Product(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		opt := Options{
			CountThreads:  2,
			CountElements: 100,
		}

		expectedList := map[string]*models.Product{
			"1": {
				ID: "1",
				Name: "Product 1",
			},
			"2": {
				ID: "2",
				Name: "Product 2",
			},
			"3": {
				ID: "3",
				Name: "Product 3",
			},
			"4": {
				ID: "4",
				Name: "Product 4",
			},
		}

		listGames := map[string]*models.ListGames{
			"1": {
				ID: "1",
				Links: []models.ListGamesLink{
					{Bucket: "test1", URL: "url 1"},
					{Bucket: "test2", URL: "url 2"},
				},
			},
			"2": {
				ID: "2",
				Links: []models.ListGamesLink{
					{Bucket: "test3", URL: "url 3"},
					{Bucket: "test4", URL: "url 4"},
				},
			},
		}

		params := models.ListParams{
			Size:  100,
			Start: 0,
		}

		cli := &MockPSNClienter{}
		cli.On("ListGames", params).
			Return(listGames["1"], nil).
			Once()

		params.Start = 100
		cli.On("ListGames", params).
			Return(listGames["2"], nil).
			Once()

		cli.On("ListGames", mock.Anything).
			Return(&models.ListGames{}, nil).
			Twice()

		cli.On("ProductByURL", listGames["1"].Links[0].URL).
			Return(expectedList["1"], nil).
			Once()

		cli.On("ProductByURL", listGames["1"].Links[1].URL).
			Return(expectedList["2"], nil).
			Once()

		cli.On("ProductByURL", listGames["2"].Links[0].URL).
			Return(expectedList["3"], nil).
			Once()

		cli.On("ProductByURL", listGames["2"].Links[1].URL).
			Return(expectedList["4"], nil).
			Once()

		p := New(opt, cli)
		ch := p.Product(context.Background())

		done := make(chan bool)
		actual := make(map[string]*models.Product)

		go func(done chan<- bool, ch <-chan ResponseProduct, actual map[string]*models.Product, t *testing.T) {
			timer := time.After(time.Second)
			for {
				select {
				case <-timer:
					done <- false
				case el, ok := <-ch:
					if !ok {
						done <- true
						return
					}
					require.NoError(t, el.Error)
					actual[el.Data.ID] = el.Data
				}
			}
		}(done, ch, actual, t)

		require.Truef(t, <-done, "Channel is not closed")
		require.Equal(t, expectedList, actual)

		cli.AssertExpectations(t)
	})

	t.Run("Context canceled", func(t *testing.T) {
		expectedList := &models.ListGames{
			ID: "1",
			Links: []models.ListGamesLink{
				{Bucket: "test"},
				{Bucket: "test"},
			},
		}

		cli := &MockPSNClienter{}
		cli.On("ListGames", mock.Anything).
			Return(expectedList, nil).
			Maybe()
		cli.On("ProductByURL", mock.Anything).
			Return(&models.Product{ID: "ID"}, nil).
			Maybe()

		opt := Options{
			CountThreads:  2,
			CountElements: 100,
		}
		p := New(opt, cli)
		ctx, cancel := context.WithCancel(context.Background())
		ch := p.Product(ctx)

		done := make(chan bool)
		var actualErr error

		go func(done chan<- bool, ch <-chan ResponseProduct, cancel context.CancelFunc) {
			timer := time.After(time.Second)
			for {
				select {
				case <-timer:
					done <- false
					return
				case el, ok := <-ch:
					if !ok {
						done <- true
						return
					}
					actualErr = el.Error
					cancel()
				}
			}
		}(done, ch, cancel)

		require.Truef(t, <-done, "Channel is not closed")
		require.ErrorIs(t, actualErr, context.Canceled)

		cli.AssertExpectations(t)
	})

	t.Run("Product error", func(t *testing.T) {
		cli := &MockPSNClienter{}
		cli.On("ListGames", mock.Anything).
			Return(
				&models.ListGames{
					ID: "1",
					Links: []models.ListGamesLink{
						{Bucket: "test"},
					},
				},
				nil,
			).
			Once()
		cli.On("ListGames", mock.Anything).
			Return(&models.ListGames{}, nil).
			Once()

		cli.On("ProductByURL", mock.Anything).
			Return(nil, errors.New("some error")).
			Once()

		opt := Options{
			CountThreads:  1,
			CountElements: 100,
		}
		p := New(opt, cli)
		ch := p.Product(context.Background())

		done := make(chan bool)

		go func(done chan<- bool, ch <-chan ResponseProduct, t *testing.T) {
			timer := time.After(time.Second)
		loop:
			for {
				select {
				case <-timer:
					done <- false
				case el, ok := <-ch:
					if !ok {
						done <- true
						break loop
					}
					require.Error(t, el.Error)
				}
			}
		}(done, ch, t)

		require.Truef(t, <-done, "Channel is not closed")

		cli.AssertExpectations(t)
	})

	t.Run("List games error", func(t *testing.T) {
		cli := &MockPSNClienter{}
		cli.On("ListGames", mock.Anything).
			Return(nil, errors.New("some error")).
			Once()

		opt := Options{
			CountThreads:  1,
			CountElements: 100,
		}
		p := New(opt, cli)
		ch := p.Product(context.Background())

		done := make(chan bool)

		go func(done chan<- bool, ch <-chan ResponseProduct, t *testing.T) {
			timer := time.After(time.Second)
		loop:
			for {
				select {
				case <-timer:
					done <- false
				case el, ok := <-ch:
					if !ok {
						done <- true
						break loop
					}
					require.Error(t, el.Error)
				}
			}
		}(done, ch, t)

		require.Truef(t, <-done, "Channel is not closed")

		cli.AssertExpectations(t)
	})
}

func TestParser_startCalculation(t *testing.T) {
	opt := Options{
		CountThreads:  3,
		CountElements: 100,
	}

	p := New(opt, nil)
	require.Equal(t, 300, p.startCalculation(0))
	require.Equal(t, 400, p.startCalculation(100))
	require.Equal(t, 600, p.startCalculation(300))
	require.Equal(t, 700, p.startCalculation(400))
}
