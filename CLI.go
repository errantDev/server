package poker

import (
	"bufio"
	"io"
	"strings"
	"time"
)

type CLI struct {
	playerstore PlayerStore
	in          *bufio.Scanner
}

type BlindAlerter interface {
	ScheduledAlertAt(duration time.Duration, amount int)
}

func NewCLI(store PlayerStore, in io.Reader, alerter BlindAlerter) *CLI {
	return &CLI{
		playerstore: store,
		in:          bufio.NewScanner(in),
	}
}

func (c *CLI) PlayPoker() {
	userInput := c.readLine()
	c.playerstore.RecordWin(extractWinner(userInput))
}

func (c *CLI) readLine() string {
	c.in.Scan()
	return c.in.Text()
}

func extractWinner(userInput string) string {
	return strings.Replace(userInput, " wins", "", 1)
}
