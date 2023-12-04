package metrics

type Metric interface {
	GetType() string
	GetName() string
	GetValue() interface{}

	SetValue(interface{})
}
