package metric

import (
	"github.com/arxon31/metrics-collector/pkg/e"
	"strconv"
)

const (
	GaugeCount   = 31
	CounterCount = 1
	GaugeType    = "gauge"
	CounterType  = "counter"
)

type MetricDTO struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

type Metrics struct {
	Gauges   map[Name]Gauge
	Counters map[Name]Counter
}

func New() *Metrics {
	return &Metrics{
		Gauges:   make(map[Name]Gauge, GaugeCount),
		Counters: make(map[Name]Counter, CounterCount),
	}
}

type Name string

const (
	Alloc           = Name("Alloc")
	BuckHashSys     = Name("BuckHashSys")
	Frees           = Name("Frees")
	GCCPUFraction   = Name("GCCPUFraction")
	GCSys           = Name("GCSys")
	HeapAlloc       = Name("HeapAlloc")
	HeapIdle        = Name("HeapIdle")
	HeapInuse       = Name("HeapInuse")
	HeapObjects     = Name("HeapObjects")
	HeapReleased    = Name("HeapReleased")
	HeapSys         = Name("HeapSys")
	LastGC          = Name("LastGC")
	Lookups         = Name("Lookups")
	MCacheInuse     = Name("MCacheInuse")
	MCacheSys       = Name("MCacheSys")
	MSpanInuse      = Name("MSpanInuse")
	MSpanSys        = Name("MSpanSys")
	Mallocs         = Name("Mallocs")
	NextGC          = Name("NextGC")
	NumForcedGC     = Name("NumForcedGC")
	NumGC           = Name("NumGC")
	OtherSys        = Name("OtherSys")
	PauseTotalNs    = Name("PauseTotalNs")
	StackInuse      = Name("StackInuse")
	StackSys        = Name("StackSys")
	Sys             = Name("Sys")
	TotalAlloc      = Name("TotalAlloc")
	RandomValue     = Name("RandomValue")
	PollCount       = Name("PollCount")
	TotalMemory     = Name("TotalMemory")
	FreeMemory      = Name("FreeMemory")
	CPUUtilization1 = Name("CPUUtilization1")
)

type Gauge float64

func (*Gauge) GaugeFromString(value string) (float64, error) {
	const op = "metric.Parse(*Gauge)"
	val, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, e.WrapError(op, "failed to parse value", err)
	}
	return val, nil
}

type Counter int64

func (*Counter) CounterFromString(value string) (int64, error) {
	const op = "metric.Parse(*Counter)"
	val, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, e.WrapError(op, "failed to parse value", err)
	}
	return val, nil
}
