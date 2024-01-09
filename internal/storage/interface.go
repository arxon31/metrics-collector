package storage

import "context"

// Пока просто проброшу контекст, ниже думаю над TODO
type Storage interface {
	Replace(ctx context.Context, name string, value float64) error
	Count(ctx context.Context, name string, value int64) error
	GaugeValue(ctx context.Context, name string) (float64, error)
	CounterValue(ctx context.Context, name string) (int64, error)
	Values(ctx context.Context) (string, error)
	ValuesJSON(ctx context.Context) (string, error)
	RestoreFromJSON(ctx context.Context, values string) error
}

// TODO: подумать над реализацией отмены транзакции
func undo(ctx context.Context) {

}
