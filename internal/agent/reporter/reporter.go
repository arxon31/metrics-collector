package reporter

import "net/http"

type Reporter interface {
	Report(requests chan *http.Request)
}

var xXxXx string
