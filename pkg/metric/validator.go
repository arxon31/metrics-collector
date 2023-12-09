package metric

type Validator interface {
	Validate(value string) (interface{}, error)
}
