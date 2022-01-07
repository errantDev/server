package poker

import "io"

type Game interface {
	Start(numberOfPlayers int, to io.Writer)
	Finish(winner string)
}
