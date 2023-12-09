package httpserver

import "net/http"

type middleware func(http.Handler) http.Handler

func Chain(h http.Handler, middlewares ...middleware) http.Handler {
	for _, m := range middlewares {
		h = m(h)
	}
	return h
}

func reqParamsCheck(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method"+r.Method+" is not allowed", http.StatusMethodNotAllowed)
			return
		}

		//if r.Header.Get("Content-Type") != "text/plain" {
		//	http.Error(w, "Only text/plain is supported", http.StatusUnprocessableEntity)
		//	return
		//
		//}
		next.ServeHTTP(w, r)
	})
}
