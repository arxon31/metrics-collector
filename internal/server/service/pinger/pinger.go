package pinger

type pinger interface {
	Ping() error
}

type pingerService struct {
	pinger pinger
}

func NewPingerService(p pinger) *pingerService {
	return &pingerService{
		pinger: p,
	}
}

// PingDB pings the database.
func (s *pingerService) PingDB() error {
	return s.pinger.Ping()
}
