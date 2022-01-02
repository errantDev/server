package poker_test

import (
	poker "server"
	"strings"
	"testing"
	"time"
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
	t.Run("Schedules printing of blind values", func(t *testing.T) {
		in := strings.NewReader("Rand wins\n")
		playerstore := &poker.StubPlayerStore{}
		blindAlerter := &SpyBlindAlerter{}
		cli := poker.NewCLI(playerstore, in, blindAlerter)
		cli.PlayPoker()
		if len(blindAlerter.alerts) != 1 {
			t.Fatal("Expected a blind alert to be scheduled")
		}
	})

}

type SpyBlindAlerter struct {
	alerts []struct {
		scheduledAt time.Duration
		amount      int
	}
}

func (s *SpyBlindAlerter) ScheduledAlertAt(duration time.Duration, amount int) {
	s.alerts = append(s.alerts, struct {
		scheduledAt time.Duration
		amount      int
	}{
		scheduledAt: duration,
		amount:      amount,
	})

}
