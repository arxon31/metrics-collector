package metrics

type gaugeMetric struct {
	Type  string
	Name  string
	Value float64
}

func NewGaugeMetric(metricType string, metricName string, metricValue float64) Metric {
	return &gaugeMetric{
		Type:  metricType,
		Name:  metricName,
		Value: metricValue,
	}
}

func (g *gaugeMetric) GetType() string {
	return g.Type
}

func (g *gaugeMetric) GetName() string {
	return g.Name
}

func (g *gaugeMetric) GetValue() interface{} {
	return g.Value
}

func (g *gaugeMetric) SetValue(value interface{}) {
	g.Value = value.(float64)
}
