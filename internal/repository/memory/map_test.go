package memory

import (
	"context"
	"github.com/arxon31/metrics-collector/internal/entity"
	"testing"
)

// BenchmarkMapStorageReplace benchmarks the Replace method of MapStorage.
func BenchmarkMapStorageReplace(b *testing.B) {
	s := NewMapStorage()
	ctx := context.Background()

	b.StopTimer()
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		s.Replace(ctx, "example", 10.0)
	}
}

// BenchmarkMapStorageCount benchmarks the Count method of MapStorage.
func BenchmarkMapStorageCount(b *testing.B) {
	s := NewMapStorage()
	ctx := context.Background()

	b.StopTimer()
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		s.Count(ctx, "example", 10)
	}
}

// BenchmarkMapStorageValues benchmarks the Values method of MapStorage.
func BenchmarkMapStorageStoreBatch(b *testing.B) {
	s := NewMapStorage()
	ctx := context.Background()
	exampleFloat := 10.0
	exampleInt := int64(10)
	b.StopTimer()

	for i := 0; i < b.N; i++ {
		b.StartTimer()
		s.StoreBatch(ctx, []entity.MetricDTO{
			{
				MetricType: "gauge",
				Name:       "example_gauge",
				Gauge:      &exampleFloat,
			},
			{
				MetricType: "counter",
				Name:       "example_counter",
				Counter:    &exampleInt,
			},
		})
	}
}
