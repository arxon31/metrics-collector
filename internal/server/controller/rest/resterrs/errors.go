package resterrs

import "errors"

var (
	ErrUnexpectedValue  = errors.New("unexpected metric value")
	ErrMetricNotFound   = errors.New("metric not found")
	ErrUnexpectedType   = errors.New("unexpected metric type")
	ErrUnexpectedFormat = errors.New("unexpected metric format")

	ErrInternalServer = errors.New("internal server error")
)
