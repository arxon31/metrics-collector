package middlewares

import (
	"net"
	"net/http"
)

const ipHeader = "X-Real-IP"

type subnetMiddleware struct {
	subnet *net.IPNet
}

func NewSubnetMiddleware(CIDR string) (*subnetMiddleware, error) {
	_, subnet, err := net.ParseCIDR(CIDR)
	if err != nil {
		return nil, err
	}

	return &subnetMiddleware{subnet: subnet}, nil

}

func (s *subnetMiddleware) WithSubnetCheck(next http.Handler) http.Handler {
	if s.subnet != nil {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			ipString := request.Header.Get(ipHeader)

			ip := net.ParseIP(ipString)

			if !s.subnet.Contains(ip) {
				http.Error(writer, "forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(writer, request)
		})
	}

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		next.ServeHTTP(writer, request)
	})
}
