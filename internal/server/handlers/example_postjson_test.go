package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/arxon31/metrics-collector/internal/repository/memory"
	"github.com/arxon31/metrics-collector/pkg/metric"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
)

var repo = memory.NewMapStorage()
var logger, _ = zap.NewDevelopment()
var sugared = logger.Sugar()

func ExamplePostJSONMetric() {

	postJSONHandler := &PostJSONMetric{Storage: repo, Provider: repo, Logger: sugared}

	exampleGauge := 10.0
	exampleCounter := int64(10)
	gaugeMetric := metric.MetricDTO{
		MType: "gauge",
		ID:    "example_gauge",
		Value: &exampleGauge,
	}
	counterMetric := metric.MetricDTO{
		MType: "counter",
		ID:    "example_counter",
		Delta: &exampleCounter,
	}

	gaugeJSON, err := json.Marshal(gaugeMetric)
	if err != nil {
		panic(err)
	}
	counterJSON, err := json.Marshal(counterMetric)
	if err != nil {
		panic(err)
	}

	reqGauge := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(gaugeJSON))
	reqCounter := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(counterJSON))
	w := httptest.NewRecorder()

	postJSONHandler.ServeHTTP(w, reqGauge)
	fmt.Println(w.Result().StatusCode)
	postJSONHandler.ServeHTTP(w, reqCounter)
	fmt.Println(w.Result().StatusCode)

	// Output:
	// 200
	// 200

}
