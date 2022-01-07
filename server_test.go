package poker

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

//unit tests
var (
	dummyPlayerStore = &StubPlayerStore{}
	dummyGame        = &GameSpy{}
)

func TestGETPlayers(t *testing.T) {
	store := StubPlayerStore{
		map[string]int{
			"Pepper": 20,
			"Floyd":  10,
		},
		nil,
		nil,
	}
	server, _ := NewPlayerServer(&store, dummyGame)
	t.Run("Returns Pepper's score", func(t *testing.T) {

		request := newPlayersRequest(http.MethodGet, "Pepper")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response, http.StatusOK)
		assertResponseBody(t, response.Body.String(), "20")
	})
	t.Run("returns Floyd's score", func(t *testing.T) {
		request := newPlayersRequest(http.MethodGet, "Floyd")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response, http.StatusOK)
		assertResponseBody(t, response.Body.String(), "10")
	})
	t.Run("returns 404 for missing players", func(t *testing.T) {
		request := newPlayersRequest(http.MethodGet, "Apollo")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response, http.StatusNotFound)
	})

}

func TestStoreWins(t *testing.T) {
	store := StubPlayerStore{
		map[string]int{},
		nil,
		nil,
	}
	server, _ := NewPlayerServer(&store, dummyGame)
	t.Run("Records wins on post", func(t *testing.T) {
		player := "Pepper"
		request := newPlayersRequest(http.MethodPost, player)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response, http.StatusAccepted)

		if len(store.winCalls) != 1 {
			t.Errorf("got %d calls to Recordwin, want %d", len(store.winCalls), 1)
		}

		if store.winCalls[0] != player {
			t.Errorf("did not store correct winner got %q want %q", store.winCalls[0], player)
		}
	})
}

func TestLeague(t *testing.T) {

	t.Run("it returns the league table as JSON", func(t *testing.T) {
		wantedLeague := []Player{
			{"Cleo", 32},
			{"Chris", 20},
			{"Tiest", 14},
		}

		store := StubPlayerStore{nil, nil, wantedLeague}
		server, _ := NewPlayerServer(&store, dummyGame)

		request := newLeagueRequest(http.MethodGet)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		got := getLeagueFromResponse(t, response.Body)
		assertStatus(t, response, http.StatusOK)
		assertLeague(t, got, wantedLeague)
		assertContentType(t, response.Result().Header.Get("content-type"))
	})
}

// Integration Tests:
func TestRecordingWinsAndRetrievingLeague(t *testing.T) {

	database, cleanDatabase := createTempFile(t, `[]`)
	defer cleanDatabase()
	store, err := NewFileSystemPlayerStore(database)
	assertNoError(t, err)

	server, _ := NewPlayerServer(store, dummyGame)
	player := "Pepper"

	server.ServeHTTP(httptest.NewRecorder(), newPlayersRequest(http.MethodPost, player))
	server.ServeHTTP(httptest.NewRecorder(), newPlayersRequest(http.MethodPost, player))
	server.ServeHTTP(httptest.NewRecorder(), newPlayersRequest(http.MethodPost, player))

	response := httptest.NewRecorder()
	server.ServeHTTP(response, newLeagueRequest(http.MethodGet))
	wantedLeague := []Player{
		{"Pepper", 3},
	}

	got := getLeagueFromResponse(t, response.Body)
	assertStatus(t, response, http.StatusOK)
	assertLeague(t, got, wantedLeague)
	assertContentType(t, response.Result().Header.Get("content-type"))
}

func TestGame(t *testing.T) {
	t.Run("GET /game returns 200", func(t *testing.T) {
		server, _ := NewPlayerServer(&StubPlayerStore{}, dummyGame)

		request := newGameRequest(http.MethodGet)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)
		assertStatus(t, response, http.StatusOK)
	})
	t.Run("start a game with 3 players and declare Paul the winner", func(t *testing.T) {
		game := &GameSpy{}
		winner := "Paul"
		server := httptest.NewServer(mustMakePlayerServer(t, dummyPlayerStore, game))
		ws := mustDialWS(t, "ws"+strings.TrimPrefix(server.URL, "http")+"/ws")

		defer server.Close()
		defer ws.Close()

		writeWSMessage(t, ws, "3")
		writeWSMessage(t, ws, winner)

		time.Sleep(10 * time.Millisecond)
		assertStartedWith(t, *game, 3)
		assertFinishedWith(t, *game, winner)
	})
}

func newPlayersRequest(method, name string) *http.Request {
	req, _ := http.NewRequest(method, fmt.Sprintf("/players/%s", name), nil)
	return req
}

func newGameRequest(method string) *http.Request {
	req, _ := http.NewRequest(method, "/league", nil)
	return req
}

func newLeagueRequest(method string) *http.Request {
	req, _ := http.NewRequest(method, "/league", nil)
	return req
}

func getLeagueFromResponse(t testing.TB, body io.Reader) (league []Player) {
	t.Helper()
	league, _ = NewLeague(body)
	return league
}

func mustDialWS(t *testing.T, url string) *websocket.Conn {
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)

	if err != nil {
		t.Fatalf("could not open a ws connection on %s %v", url, err)
	}

	return ws
}

func writeWSMessage(t testing.TB, conn *websocket.Conn, message string) {
	t.Helper()
	if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
		t.Fatalf("could not send message over ws connection %v", err)
	}
}

func mustMakePlayerServer(t *testing.T, store PlayerStore, game Game) *PlayerServer {
	server, err := NewPlayerServer(store, game)
	if err != nil {
		t.Fatal("problem creating player server", err)
	}
	return server
}

func assertResponseBody(t testing.TB, got, want string) {
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func assertStartedWith(t testing.TB, game GameSpy, want int) {
	if game.StartedWith != want {
		t.Errorf("Wanted start with of %d but got %d", want, game.StartedWith)
	}
}

func assertFinishedWith(t testing.TB, game GameSpy, want string) {
	if game.FinishedWith != want {
		t.Errorf("Wanted winner of %v but got %v", want, game.FinishedWith)
	}
}

func assertStatus(t testing.TB, got *httptest.ResponseRecorder, want int) {
	if got.Code != want {
		t.Errorf("did not get correct status. Got %d, want %d", got.Code, want)
	}
}

func assertPlayerScore(t testing.TB, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}

func assertLeague(t testing.TB, got, want []Player) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

func assertContentType(t testing.TB, got string) {
	if got != jsonContentType {
		t.Errorf("response did not have content type application/json, got %v", got)
	}
}

func assertNoError(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("didn't expect error, but received %v", err)
	}
}
