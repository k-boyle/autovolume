package plex

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

const playingMessage = `{"NotificationContainer": {"type": "playing", "PlaySessionStateNotification": [
	{"clientIdentifier": "player", "ratingKey": "42", "state": "playing"}
]}}`

func TestWebsocketPublishesPlayingNotifications(t *testing.T) {
	upgrader := websocket.Upgrader{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/:/websockets/notifications" || r.Header.Get("X-Plex-Token") != "token" {
			http.Error(w, "unexpected request", http.StatusBadRequest)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		if err := conn.WriteMessage(websocket.TextMessage, []byte(playingMessage)); err != nil {
			return
		}
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}))
	t.Cleanup(srv.Close)

	ws := NewWebsocket(t.Context())
	t.Cleanup(func() { _ = ws.Close() })

	subscription, err := ws.Subscribe()
	if err != nil {
		t.Fatal(err)
	}

	details := &WebsocketDetails{Url: mustParse(srv.URL), Token: "token", ClientId: "player"}
	if err := ws.ServerChanged(details); err != nil {
		t.Fatal(err)
	}

	select {
	case notification := <-subscription.Read():
		if notification.RatingKey != "42" || notification.State != "playing" {
			t.Errorf("notification = %+v, want rating key 42 playing", notification)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for notification")
	}
}
