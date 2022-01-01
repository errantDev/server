package poker

import (
	"bufio"
	"io"
	"strings"
)

type CLI struct {
	playerstore PlayerStore
	in          io.Reader
}

func (c *CLI) PlayPoker() {
	reader := bufio.NewScanner(c.in)
	reader.Scan()
	c.playerstore.RecordWin(extractWinner(reader.Text()))
}

func extractWinner(userInput string) string {
	return strings.Replace(userInput, " wins", "", 1)
}
