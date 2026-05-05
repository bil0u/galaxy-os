package discord

import (
	"github.com/disgoorg/paginator"
)

var Paginator *paginator.Manager

func InitPaginator() *paginator.Manager {
	Paginator = paginator.New()
	return Paginator
}
