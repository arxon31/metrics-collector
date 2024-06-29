package entity

import (
	"fmt"
	"strconv"
)

const (
	GaugeCount   = 31
	CounterCount = 1
	GaugeType    = "gauge"
	CounterType  = "counter"
)

type Gauge float64
type Counter int64

const (
	Alloc           = "Alloc"
	BuckHashSys     = "BuckHashSys"
	Frees           = "Frees"
	GCCPUFraction   = "GCCPUFraction"
	GCSys           = "GCSys"
	HeapAlloc       = "HeapAlloc"
	HeapIdle        = "HeapIdle"
	HeapInuse       = "HeapInuse"
	HeapObjects     = "HeapObjects"
	HeapReleased    = "HeapReleased"
	HeapSys         = "HeapSys"
	LastGC          = "LastGC"
	Lookups         = "Lookups"
	MCacheInuse     = "MCacheInuse"
	MCacheSys       = "MCacheSys"
	MSpanInuse      = "MSpanInuse"
	MSpanSys        = "MSpanSys"
	Mallocs         = "Mallocs"
	NextGC          = "NextGC"
	NumForcedGC     = "NumForcedGC"
	NumGC           = "NumGC"
	OtherSys        = "OtherSys"
	PauseTotalNs    = "PauseTotalNs"
	StackInuse      = "StackInuse"
	StackSys        = "StackSys"
	Sys             = "Sys"
	TotalAlloc      = "TotalAlloc"
	RandomValue     = "RandomValue"
	PollCount       = "PollCount"
	TotalMemory     = "TotalMemory"
	FreeMemory      = "FreeMemory"
	CPUUtilization1 = "CPUUtilization1"
)

func (*Gauge) GaugeFromString(value string) (float64, error) {
	val, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("falied to parse value: %w", err)
	}
	return val, nil
}

func (*Counter) CounterFromString(value string) (int64, error) {
	val, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("falied to parse value: %w", err)
	}
	return val, nil
}
