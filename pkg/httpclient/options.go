package httpclient

import "github.com/go-resty/resty/v2"

type Option func(c *client)

func (c *client) WithRetries(count int) {
	c.client.RetryCount = count
}

func (c *client) WithWorkers(count int) {
	c.numWorkers = count
}

func conditionWithUnsuccessfulResponse(response *resty.Response, err error) bool {
	if !response.IsSuccess() {
		return true
	}
	return false
}
