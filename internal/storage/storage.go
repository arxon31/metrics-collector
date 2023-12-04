package storage

type Storage interface {
	Update(name string, newVal interface{}) error
	GetMetric(name string) interface{}

	IsExist(name string) bool
}
