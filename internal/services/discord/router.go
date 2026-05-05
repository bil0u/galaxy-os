package discord

import (
	"github.com/disgoorg/disgo/handler"
)

var Router *handler.Mux

func InitRouter() *handler.Mux {
	Router = handler.New()
	return Router
}
