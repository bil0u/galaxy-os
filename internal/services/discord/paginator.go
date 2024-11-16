package discord

import (
	"sync"

	"github.com/disgoorg/paginator"
)

var (
	Paginator *paginator.Manager
)

func InitPaginator() *paginator.Manager {
	sync.OnceFunc(func() {
		Paginator = paginator.New()
	})()

	return Paginator
}
