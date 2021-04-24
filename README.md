# Playstation store parser

Парсинг игр с playstation store.

Вызов функции `Catalog` или `Product` вернет канал, в котором содержаться данные от playstation store

## Установка
```shell script
go get github.com/EvgenyGavrilov/psnparser
```

## Пример использования
**Парсинг игр**
```go
package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/EvgenyGavrilov/psnclient"
	"github.com/EvgenyGavrilov/psnclient/models"
	"github.com/EvgenyGavrilov/psnparser"
)

func main() {
	cli, err := psnclient.New("ru", "ru")
	if err != nil {
		panic(err)
	}

	opt := psnparser.Options{
		CountThreads:  2,
		CountElements: 10,
	}

	p := psnparser.New(opt, cli)

	ctx, cancel := context.WithCancel(context.Background())
	ch := p.Product(ctx)
	for el := range ch {
		if el.Error != nil && el.Error != context.Canceled {
			err, ok := el.Error.(*models.HTTPError)
			if ok && err.StatusCode == http.StatusNoContent {
				continue
			}
			panic(el.Error)
		}
		if el.Data != nil {
			fmt.Println(el.Data.ID, "-", el.Data.Name)
			cancel()
		}
	}
}
```

В результате получим следующие:
```text
EP1018-PPSA01696_00-BACK4BLOOD000000 - Back 4 Blood: Стандартное издание PS4 and PS5
EP1018-PPSA01696_00-BACK4BLOODDELUXE - Back 4 Blood: Deluxe-издание PS4 and PS5
```
