package metrics

type counterMetric struct {
	Type  string
	Name  string
	Value int64
}

func NewCounterMetric(metricType string, metricName string, metricValue int64) Metric {
	return &counterMetric{
		Type:  metricType,
		Name:  metricName,
		Value: metricValue,
	}
}

func (c *counterMetric) GetType() string {
	return c.Type
}

func (c *counterMetric) GetName() string {
	return c.Name
}

func (c *counterMetric) GetValue() interface{} {
	return c.Value
}

func (c *counterMetric) SetValue(value interface{}) {
	c.Value = value.(int64)
}
