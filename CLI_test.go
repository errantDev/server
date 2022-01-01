package poker_test

import (
	"strings"
	"testing"
)

func TestCLI(t *testing.T) {
	t.Run("Paul's wins are recorded", func(t *testing.T) {
		in := strings.NewReader("Paul wins\n")
		playerstore := &poker.StubPlayerStore{}
		cli := &poker.CLI{playerstore, in}
		cli.PlayPoker()
		poker.AssertPlayerWin(t, playerstore, "Paul")
	})
	t.Run("Rand's wins are recorded", func(t *testing.T) {
		in := strings.NewReader("Rand wins\n")
		playerstore := &poker.StubPlayerStore{}
		cli := &poker.CLI{playerstore, in}
		cli.PlayPoker()
		poker.AssertPlayerWin(t, playerstore, "Rand")
	})
}
