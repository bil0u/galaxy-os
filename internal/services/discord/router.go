package discord

import (
	"github.com/disgoorg/disgo/handler"
)

var router *handler.Mux

func InitRouter() *handler.Mux {
	router = handler.New()
	return router
}

func Router() *handler.Mux {
	return router
}
