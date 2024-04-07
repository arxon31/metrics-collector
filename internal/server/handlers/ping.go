package handlers

import (
	"net/http"
)

type Ping CustomHandler

// Ping implements http.Handler which checks connection to DB
func (h *Ping) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.Pinger.Ping()
	if err != nil {
		http.Error(w, "can not connect to DB", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
