package httpclient

import "github.com/go-resty/resty/v2"

type Option func(c *client)

func WithRetries(count int) Option {
	return func(c *client) {
		c.client.RetryCount = count
	}
}

func conditionWithUnsuccessfulResponse(response *resty.Response, err error) bool {
	if !response.IsSuccess() {
		return true
	}
	return false
}
