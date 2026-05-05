package discord

import (
	"github.com/disgoorg/paginator"
)

var pgn *paginator.Manager

func InitPaginator() *paginator.Manager {
	pgn = paginator.New()
	return pgn
}

func Paginator() *paginator.Manager {
	return pgn
}
