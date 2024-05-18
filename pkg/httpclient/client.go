package httpclient

import (
	"context"
	"github.com/go-resty/resty/v2"
)

const (
	_defaultRetryCount = 3
)

type client struct {
	client     *resty.Client
	retryCount int
	numWorkers int
}

func NewClient(opts ...Option) *client {
	c := &client{
		client:     resty.New(),
		retryCount: _defaultRetryCount,
	}

	for _, opt := range opts {
		opt(c)
	}

	c.client.SetRetryCount(c.retryCount)
	c.client.AddRetryCondition(conditionWithUnsuccessfulResponse)

	return c
}

func (c *client) DoCtx(ctx context.Context, req *resty.Request) (*resty.Response, error) {
	return req.SetContext(ctx).Send()
}
