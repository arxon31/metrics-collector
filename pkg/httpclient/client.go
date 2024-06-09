package httpclient

import (
	"net/http"
)

type client struct {
	client *http.Client
}

func NewClient() *client {
	return &client{
		client: &http.Client{},
	}
}

func (c *client) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}
