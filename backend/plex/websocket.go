package plex

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type WebSocket struct {
	cancel            context.CancelFunc
	wg                *sync.WaitGroup
	conn              *websocket.Conn
	notificationQueue chan *PlaySessionStateNotification
	msgQueue          chan *websocketMessage
	clientId          string
}

func Connect(ctx context.Context, serverUrl, serverToken, clientId string) (*WebSocket, error) {
	slog.Info("Starting WebSocket listener")

	ctx, cancel := context.WithCancel(ctx)

	base, err := url.Parse(serverUrl)
	if err != nil {
		cancel()
		return nil, err
	}
	base.Scheme = "ws"
	endpoint := base.JoinPath("/:/websockets/notifications")

	conn, _, err := websocket.DefaultDialer.Dial(endpoint.String(), http.Header{"X-Plex-Token": []string{serverToken}})
	if err != nil {
		cancel()
		return nil, err
	}

	ws := &WebSocket{
		cancel:            cancel,
		wg:                new(sync.WaitGroup),
		conn:              conn,
		notificationQueue: make(chan *PlaySessionStateNotification),
		msgQueue:          make(chan *websocketMessage),
		clientId:          clientId,
	}

	ws.wg.Go(func() {
		ws.read(ctx)
	})

	ws.wg.Go(func() {
		ws.publish(ctx)
	})

	return ws, nil
}

func (ws *WebSocket) Read() <-chan *PlaySessionStateNotification {
	return ws.notificationQueue
}

func (ws *WebSocket) Close() error {
	ws.cancel()

	writeErr := ws.conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(time.Second))
	closeErr := ws.conn.Close()

	ws.wg.Wait()
	return errors.Join(writeErr, closeErr)
}

func (ws *WebSocket) read(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		_, raw, err := ws.conn.ReadMessage()
		if err != nil {
			if ctx.Err() == nil {
				slog.Error("websocket read error:", slog.Any("error", err))
			}
			ws.cancel()
			return
		}

		msg := &websocketMessage{}
		if err := json.Unmarshal(raw, msg); err != nil {
			slog.Error("websocket unmarshal error:", slog.Any("error", err))
			continue
		}

		select {
		case ws.msgQueue <- msg:
		case <-ctx.Done():
			return
		}
	}
}

func (ws *WebSocket) publish(ctx context.Context) {
	defer close(ws.notificationQueue)

	var lastNotification *PlaySessionStateNotification
	ticker := time.NewTicker(30 * time.Second)

	for {

		select {
		case msg := <-ws.msgQueue:
			container := msg.NotificationContainer
			if container.Type != "playing" {
				continue
			}

			for _, notification := range container.PlaySessionStateNotification {
				if notification.ClientIdentifier != ws.clientId {
					continue
				}

				select {
				case ws.notificationQueue <- notification:
					lastNotification = notification
				case <-ctx.Done():
					return
				}
			}

			slog.Info("Published notifications", slog.Int("notifications", len(container.PlaySessionStateNotification)))
		case <-ticker.C:
			if lastNotification == nil {
				continue
			}

			select {
			case ws.notificationQueue <- lastNotification:
			case <-ctx.Done():
				return
			default:
			}
		case <-ctx.Done():
			return
		}
	}
}
