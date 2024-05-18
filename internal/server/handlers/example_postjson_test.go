package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/arxon31/metrics-collector/internal/entity"
	"github.com/arxon31/metrics-collector/internal/repository/memory"
	"net/http"
	"net/http/httptest"

	"go.uber.org/zap"
)

var repo = memory.NewMapStorage()
var logger, _ = zap.NewDevelopment()
var sugared = logger.Sugar()

func ExamplePostJSONMetric() {

	postJSONHandler := &PostJSONMetric{Storage: repo, Provider: repo, Logger: sugared}

	exampleGauge := 10.0
	exampleCounter := int64(10)
	gaugeMetric := entity.MetricDTO{
		MetricType: "gauge",
		Name:       "example_gauge",
		Gauge:      &exampleGauge,
	}
	counterMetric := entity.MetricDTO{
		MetricType: "counter",
		Name:       "example_counter",
		Counter:    &exampleCounter,
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
