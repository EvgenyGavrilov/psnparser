package psnparser

// Options опции парсера
type Options struct {
	// CountThreads количество одновременно работающих потоков парсера.
	CountThreads  int
	// CountElements количество запрашиваемых элементов в одном потоке.
	CountElements int
}
