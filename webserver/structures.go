package webserver

import (
	"wireguard_api/usecases"
)

type Server struct {
	Service       *usecases.Usecases
	AlloweSubnets []string
}
