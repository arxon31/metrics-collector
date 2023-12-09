package storage

import "context"

// Пока просто проброшу контекст, ниже думаю над TODO
type Storage interface {
	Replace(ctx context.Context, name string, value float64) error
	Count(ctx context.Context, name string, value int64) error
}

// TODO: подумать над реализацией отмены транзакции
func undo(ctx context.Context) {

}
