package server

import "simple-load-balancer/backend"

type ServerPool struct {
	Backends []*backend.Backend
	Current  uint64
}
