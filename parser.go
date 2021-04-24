package psnparser

import (
	"context"
	"sync"

	"github.com/EvgenyGavrilov/psnclient/models"
)

// PSNClienter интерфейс playstation API клиента.
type PSNClienter interface {
	ListGames(params models.ListParams) (*models.ListGames, error)
	ProductByURL(u string) (*models.Product, error)
}

// ResponseCatalog структура содержащая информацию о каталоге.
type ResponseCatalog struct {
	// Data ответ от playstation store.
	Data *models.ListGames
	// Error ошибка возникшая во время парсинга.
	Error error
}

// ResponseProduct  структура содержащая информацию о игре
type ResponseProduct struct {
	// Data ответ от playstation store.
	Data *models.Product
	// Error ошибка возникшая во время парсинга.
	Error error
}

// Parser парсер playstation store.
type Parser struct {
	cli PSNClienter
	opt Options
}

// New вернет новый инстанс Parser
func New(opt Options, cli PSNClienter) *Parser {
	return &Parser{
		opt: opt,
		cli: cli,
	}
}

// Catalog запуск парсинга каталога в один и более потоков в зависимости от настройки.
// Вернет канал содержащий результат ответа от playstation store и ошибку если она возникла.
// Функция принимает Context. Context может быть использован для принудительной остановки парсинга.
// Будьте внимательны в результате вызыва функции cancel context.WithCancel, парсер вернет ошибку context.Canceled.
// Ошибки такого рода обрабатываются на вашей стороне.
func (p *Parser) Catalog(ctx context.Context) <-chan ResponseCatalog {
	ch := make(chan ResponseCatalog)
	wg := &sync.WaitGroup{}
	wg.Add(p.opt.CountThreads)
	for i := 0; i < p.opt.CountThreads; i++ {
		go func(ctx context.Context, i int, wg *sync.WaitGroup) {
			p.runParseCatalog(ctx, i, ch)
			wg.Done()
		}(ctx, i, wg)
	}

	go func(ch chan ResponseCatalog, wg *sync.WaitGroup) {
		wg.Wait()
		close(ch)
	}(ch, wg)

	return ch
}

// Product запуск парсинга игр в один и более потоков в зависимости от настройки.
// Вернет канал содержащий результат ответа от playstation store и ошибку если она возникла.
// Функция принимает Context. Context может быть использован для принудительной остановки парсинга.
// Будьте внимательны в результате вызыва функции cancel context.WithCancel, парсер вернет ошибку context.Canceled.
// Ошибки такого рода обрабатываются на вашей стороне.
// TODO: При запросе игры, может вернутся ошибка с http статусом 204 и пустое тело ответа.
//   Есть подазрения, что такая игра как-то связана с другой. Скорей всего это одна и та же игра, но под разные платформы.
func (p *Parser) Product(ctx context.Context) <-chan ResponseProduct {
	ch := make(chan ResponseProduct)
	chList := p.Catalog(ctx)
	wg := &sync.WaitGroup{}
	wg.Add(p.opt.CountThreads)
	for i := 0; i < p.opt.CountThreads; i++ {
		go func(ctx context.Context, chList <-chan ResponseCatalog, chProduct chan ResponseProduct) {
			p.runParseProduct(ctx, chList, chProduct)
			wg.Done()
		}(ctx, chList, ch)
	}

	go func(ch chan ResponseProduct, wg *sync.WaitGroup) {
		wg.Wait()
		close(ch)
	}(ch, wg)

	return ch
}

func (p *Parser) runParseProduct(ctx context.Context, chList <-chan ResponseCatalog, chProduct chan ResponseProduct) {
	product := ResponseProduct{}
	for list := range chList {
		if list.Error != nil {
			product.Error = list.Error
			chProduct <- product
			continue
		}
		for _, el := range list.Data.Links {
			select {
			case <-ctx.Done():
				err := ctx.Err()
				if err != nil {
					chProduct <- ResponseProduct{
						Error: err,
					}
				}
				return
			default:
				product.Data, product.Error = p.cli.ProductByURL(el.URL)
				chProduct <- product
			}
		}
	}
}

func (p *Parser) runParseCatalog(ctx context.Context, nP int, ch chan ResponseCatalog) {
	start := nP * p.opt.CountElements
	for {
		select {
		case <-ctx.Done():
			err := ctx.Err()
			if err != nil {
				ch <- ResponseCatalog{
					Error: err,
				}
			}
			return

		default:
			params := models.ListParams{Start: start, Size: p.opt.CountElements}
			listGames, err := p.cli.ListGames(params)
			if err != nil {
				ch <- ResponseCatalog{
					Data:  listGames,
					Error: err,
				}
				return
			}

			if len(listGames.Links) == 0 {
				return
			}

			ch <- ResponseCatalog{
				Data:  listGames,
				Error: err,
			}

			start = p.startCalculation(start)
		}
	}
}

func (p *Parser) startCalculation(start int) int {
	return start + p.opt.CountThreads*p.opt.CountElements
}
