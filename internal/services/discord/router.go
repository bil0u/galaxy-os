package discord

import (
	"sync"

	"github.com/disgoorg/disgo/handler"
)

var (
	Router *handler.Mux
)

func InitRouter() *handler.Mux {
	sync.OnceFunc(func() {
		Router = handler.New()
	})()

	return Router
}
