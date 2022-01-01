package poker_test

import (
	poker "server"
	"strings"
	"testing"
)

func TestCLI(t *testing.T) {
	t.Run("Paul's wins are recorded", func(t *testing.T) {
		in := strings.NewReader("Paul wins\n")
		playerstore := &poker.StubPlayerStore{}
		cli := poker.NewCLI(playerstore, in)
		cli.PlayPoker()
		poker.AssertPlayerWin(t, playerstore, "Paul")
	})
	t.Run("Rand's wins are recorded", func(t *testing.T) {
		in := strings.NewReader("Rand wins\n")
		playerstore := &poker.StubPlayerStore{}
		cli := poker.NewCLI(playerstore, in)
		cli.PlayPoker()
		poker.AssertPlayerWin(t, playerstore, "Rand")
	})
}
