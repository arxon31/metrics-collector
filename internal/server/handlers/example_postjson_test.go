package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"go.uber.org/zap"

	"github.com/arxon31/metrics-collector/internal/repository/memory"
	"github.com/arxon31/metrics-collector/pkg/metric"
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
	defer reqGauge.Body.Close()
	reqCounter := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(counterJSON))
	defer reqCounter.Body.Close()

	wG := httptest.NewRecorder()
	wC := httptest.NewRecorder()

	postJSONHandler.ServeHTTP(wG, reqGauge)
	resGauge := wG.Result()
	defer resGauge.Body.Close()
	postJSONHandler.ServeHTTP(wC, reqCounter)
	resCounter := wC.Result()
	defer resCounter.Body.Close()

	fmt.Println(resGauge.StatusCode)
	fmt.Println(resCounter.StatusCode)

	// Output:
	// 200
	// 200

}
