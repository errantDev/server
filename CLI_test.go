package poker_test

import (
	"bytes"
	"fmt"
	poker "server"
	"strings"
	"testing"
	"time"
)

var dummyBlindAlerter = &SpyBlindAlerter{}
var dummyPlayerStore = &poker.StubPlayerStore{}
var dummyStdIn = &bytes.Buffer{}
var dummyStdOut = &bytes.Buffer{}

func TestCLI(t *testing.T) {
	t.Run("Paul's wins are recorded", func(t *testing.T) {
		in := strings.NewReader("7\nPaul wins\n")
		playerstore := &poker.StubPlayerStore{}
		dummyAlerter := &SpyBlindAlerter{}
		game := poker.NewTexasHoldem(dummyAlerter, playerstore)
		cli := poker.NewCLI(in, dummyStdOut, game)
		cli.PlayPoker()
		poker.AssertPlayerWin(t, playerstore, "Paul")
	})
	t.Run("Rand's wins are recorded", func(t *testing.T) {
		in := strings.NewReader("7\nRand wins\n")
		playerstore := &poker.StubPlayerStore{}
		dummyAlerter := &SpyBlindAlerter{}
		game := poker.NewTexasHoldem(dummyAlerter, playerstore)
		cli := poker.NewCLI(in, dummyStdOut, game)
		cli.PlayPoker()
		poker.AssertPlayerWin(t, playerstore, "Rand")
	})
	t.Run("it schedules printing of blind values", func(t *testing.T) {
		in := strings.NewReader("5\nChris wins\n")
		playerStore := &poker.StubPlayerStore{}
		blindAlerter := &SpyBlindAlerter{}
		game := poker.NewTexasHoldem(blindAlerter, playerStore)

		cli := poker.NewCLI(in, dummyStdOut, game)
		cli.PlayPoker()

		cases := []scheduledAlert{
			{0 * time.Second, 100},
			{10 * time.Minute, 200},
			{20 * time.Minute, 300},
			{30 * time.Minute, 400},
			{40 * time.Minute, 500},
			{50 * time.Minute, 600},
			{60 * time.Minute, 800},
			{70 * time.Minute, 1000},
			{80 * time.Minute, 2000},
			{90 * time.Minute, 4000},
			{100 * time.Minute, 8000},
		}

		for i, c := range cases {
			t.Run(fmt.Sprintf("%d scheduled for %v", c.amount, c.scheduledAt), func(t *testing.T) {

				if len(blindAlerter.alerts) <= i {
					t.Fatalf("alert %d was not scheduled %v", i, blindAlerter.alerts)
				}

				alert := blindAlerter.alerts[i]
				assertScheduledAlert(t, alert, c)
			})
		}
	})
	t.Run("it prompts the user to enter the number of players", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		in := strings.NewReader("7\n")

		game := &GameSpy{}
		cli := poker.NewCLI(in, stdout, game)
		cli.PlayPoker()

		assertMessageSentToUser(t, stdout, poker.PlayerPrompt)
		assertStartedWith(t, game.StartedWith, 7)
	})
}
func TestGame_Start(t *testing.T) {
	t.Run("schedules alerts on game start for 5 players", func(t *testing.T) {
		blindAlerter := &SpyBlindAlerter{}
		game := poker.NewTexasHoldem(blindAlerter, dummyPlayerStore)

		game.Start(5)

		cases := []scheduledAlert{
			{0 * time.Second, 100},
			{10 * time.Minute, 200},
			{20 * time.Minute, 300},
			{30 * time.Minute, 400},
			{40 * time.Minute, 500},
			{50 * time.Minute, 600},
			{60 * time.Minute, 800},
			{70 * time.Minute, 1000},
			{80 * time.Minute, 2000},
			{90 * time.Minute, 4000},
			{100 * time.Minute, 8000},
		}

		checkSchedulingCases(t, cases, blindAlerter)
	})
	t.Run("schedules alerts on games with 7 players", func(t *testing.T) {
		blindAlerter := &SpyBlindAlerter{}
		game := poker.NewTexasHoldem(blindAlerter, dummyPlayerStore)

		game.Start(7)

		cases := []scheduledAlert{
			{0 * time.Second, 100},
			{12 * time.Minute, 200},
			{24 * time.Minute, 300},
			{36 * time.Minute, 400},
		}
		checkSchedulingCases(t, cases, blindAlerter)
	})
	t.Run("prints error when nonnumeric value is entered and does not start game", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		in := strings.NewReader("Pies\n")
		game := &GameSpy{}

		cli := poker.NewCLI(in, stdout, game)
		cli.PlayPoker()
		if game.StartCalled {
			t.Errorf("Game should not have started")
		}

		assertMessageSentToUser(t, stdout, poker.PlayerPrompt, poker.BadPlayerInputErrMsg)
	})
}

func TestGame_Finish(t *testing.T) {
	game := &GameSpy{}

	winner := "Whiskyjack"

	game.Finish(winner)
	// poker.AssertPlayerWin(t, store, winner)
	if game.FinishedWith != winner {
		t.Errorf("wanted to player %v to win. Got %v", winner, game.FinishedWith)
	}
}

type SpyBlindAlerter struct {
	alerts []scheduledAlert
}

type scheduledAlert struct {
	scheduledAt time.Duration
	amount      int
}

type GameSpy struct {
	StartCalled  bool
	StartedWith  int
	FinishedWith string
}

func (g *GameSpy) Start(numberOfPlayers int) {
	g.StartCalled = true
	g.StartedWith = numberOfPlayers
}
func (g *GameSpy) Finish(winner string) {
	g.FinishedWith = winner
}

func (s *scheduledAlert) String() string {
	return fmt.Sprintf("%d chips at %v", s.amount, s.scheduledAt)
}

func (s *SpyBlindAlerter) ScheduledAlertAt(duration time.Duration, amount int) {
	s.alerts = append(s.alerts, scheduledAlert{duration, amount})
}

func checkSchedulingCases(t *testing.T, cases []scheduledAlert, blindAlerter *SpyBlindAlerter) {
	for i, c := range cases {
		t.Run(fmt.Sprintf("%d scheduled for %v", c.amount, c.scheduledAt), func(t *testing.T) {

			if len(blindAlerter.alerts) <= i {
				t.Fatalf("alert %d was not scheduled %v", i, blindAlerter.alerts)
			}

			alert := blindAlerter.alerts[i]
			assertScheduledAlert(t, alert, c)
		})
	}
}

func assertScheduledAlert(t testing.TB, got, want scheduledAlert) {
	if got.amount != want.amount {
		t.Errorf("got amount %d, want %d", got.amount, want.amount)
	}

	if got.scheduledAt != want.scheduledAt {
		t.Errorf("got scheduled time of %v, want %v", got.scheduledAt, want.scheduledAt)
	}
}

func assertStartedWith(t testing.TB, got, want int) {
	if got != want {
		t.Errorf("Wanted start with of %d but got %d", want, got)
	}
}

func assertMessageSentToUser(t testing.TB, stdout *bytes.Buffer, messages ...string) {
	t.Helper()
	want := strings.Join(messages, "")
	got := stdout.String()
	if got != want {
		t.Errorf("got %q sent to stdout but expected %+v", got, want)
	}

}
