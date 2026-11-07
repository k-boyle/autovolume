package plex

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var closed = errors.New("websocket server closed")

type Websocket struct {
	cancel           context.CancelFunc
	changeMu         sync.Mutex
	connMu           sync.Mutex
	subMu            sync.Mutex
	conn             *websocket.Conn
	changeRequest    chan *changeRequest
	msgQueue         chan *websocketMessage
	done             <-chan struct{}
	subscribers      map[string]*Subscription
	clientId         string
	lastNotification *PlaySessionStateNotification
}

type Subscription struct {
	id string
	ch chan *PlaySessionStateNotification
}

type changeRequest struct {
	details *WebsocketDetails
	ch      chan error
}

type WebsocketDetails struct {
	Url      *url.URL
	Token    string
	ClientId string
}

func NewWebsocket(ctx context.Context) *Websocket {
	ctx, cancel := context.WithCancel(ctx)

	ws := &Websocket{
		cancel:        cancel,
		changeRequest: make(chan *changeRequest),
		msgQueue:      make(chan *websocketMessage),
		done:          ctx.Done(),
		subscribers:   make(map[string]*Subscription),
	}

	go ws.read(ctx)
	go ws.publish()

	return ws
}

func (ws *Websocket) ServerChanged(details *WebsocketDetails) error {
	select {
	case <-ws.done:
		return closed
	default:
	}

	ws.changeMu.Lock()
	defer ws.changeMu.Unlock()

	err := ws.tryClose()

	request := &changeRequest{
		details: details,
		ch:      make(chan error, 1),
	}

	select {
	case ws.changeRequest <- request:
		select {
		case connErr := <-request.ch:
			return errors.Join(connErr, err)
		case <-ws.done:
			return closed
		}
	case <-ws.done:
		return closed
	}
}

func (ws *Websocket) read(ctx context.Context) {
	defer func() {
		if err := ws.tryClose(); err != nil {
			slog.Error("failed to close websocket", slog.Any("error", err))
		}
	}()

	var lastChange *WebsocketDetails

	for {
		if ws.conn == nil {
			select {
			case request := <-ws.changeRequest:
				lastChange = request.details
				err := ws.connect(ctx, request.details)
				request.ch <- err
				if err != nil {
					continue
				}
			case <-time.After(time.Second * 5):
				if lastChange == nil {
					continue
				}

				slog.Info("timed out waiting for websocket change, reconnecting...")
				if err := ws.connect(ctx, lastChange); err != nil {
					slog.Error("failed to reconnect to websocket", slog.Any("error", err))
					continue
				}
			case <-ws.done:
				return
			}
		}

		select {
		case request := <-ws.changeRequest:
			lastChange = request.details
			err := ws.connect(ctx, request.details)
			request.ch <- err
			if err != nil {
				continue
			}
		case <-ws.done:
			return
		default:
		}

		_, raw, err := ws.conn.ReadMessage()
		if err != nil {
			ws.connMu.Lock()
			ws.conn = nil
			ws.connMu.Unlock()

			if !errors.Is(err, net.ErrClosed) {
				slog.Error("websocket read error:", slog.Any("error", err))
			}
			continue
		}

		msg := &websocketMessage{}
		if err := json.Unmarshal(raw, msg); err != nil {
			slog.Error("websocket unmarshal error:", slog.Any("error", err))
			continue
		}

		select {
		case ws.msgQueue <- msg:
		case <-ws.done:
			return
		}
	}
}

func (ws *Websocket) connect(ctx context.Context, details *WebsocketDetails) error {
	if err := ws.tryClose(); err != nil && !errors.Is(err, net.ErrClosed) {
		slog.Error("failed to close websocket", slog.Any("error", err))
	}

	url := details.Url.JoinPath("/:/websockets/notifications")
	url.Scheme = "ws"
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, url.String(), http.Header{"X-Plex-Token": []string{details.Token}})
	if err != nil {
		return err
	}

	ws.connMu.Lock()
	ws.conn = conn
	ws.clientId = details.ClientId
	ws.connMu.Unlock()

	return nil
}

func (ws *Websocket) publish() {
	defer func() {
		ws.subMu.Lock()
		defer ws.subMu.Unlock()

		for _, subscriber := range ws.subscribers {
			close(subscriber.ch)
		}

		ws.subscribers = nil
	}()

	ticker := time.NewTicker(30 * time.Second)

	for {
		select {
		case msg := <-ws.msgQueue:
			container := msg.NotificationContainer
			if container.Type != "playing" {
				continue
			}

			ws.connMu.Lock()
			clientId := ws.clientId
			ws.connMu.Unlock()

			for _, notification := range container.PlaySessionStateNotification {
				if notification.ClientIdentifier != clientId {
					continue
				}

				ws.subMu.Lock()
				ws.lastNotification = notification
				for _, subscriber := range ws.subscribers {
					select {
					case subscriber.ch <- notification:
					default:
					}
				}
				ws.subMu.Unlock()
			}

			slog.Debug("Published notifications", slog.Int("notifications", len(container.PlaySessionStateNotification)))
		case <-ticker.C:
			if ws.lastNotification == nil {
				continue
			}

			ws.connMu.Lock()
			clientId := ws.clientId
			ws.connMu.Unlock()

			if ws.lastNotification.ClientIdentifier != clientId {
				continue
			}

			ws.subMu.Lock()
			for _, subscriber := range ws.subscribers {
				select {
				case subscriber.ch <- ws.lastNotification:
				default:
				}
			}
			ws.subMu.Unlock()
		case <-ws.done:
			return
		}
	}
}

func (ws *Websocket) Subscribe() (*Subscription, error) {
	ws.subMu.Lock()
	defer ws.subMu.Unlock()

	select {
	case <-ws.done:
		return nil, closed
	default:
	}

	subscription := &Subscription{
		id: rand.Text(),
		ch: make(chan *PlaySessionStateNotification, 10),
	}

	ws.subscribers[subscription.id] = subscription

	if ws.lastNotification == nil {
		return subscription, nil
	}

	ws.connMu.Lock()
	clientId := ws.clientId
	ws.connMu.Unlock()

	if ws.lastNotification.ClientIdentifier == clientId {
		subscription.ch <- ws.lastNotification
	}

	return subscription, nil
}

func (ws *Websocket) Unsubscribe(subscription *Subscription) {
	ws.subMu.Lock()
	defer ws.subMu.Unlock()

	if _, ok := ws.subscribers[subscription.id]; ok {
		delete(ws.subscribers, subscription.id)
		close(subscription.ch)
	}
}

func (s *Subscription) Read() <-chan *PlaySessionStateNotification {
	return s.ch
}

func (ws *Websocket) Close() error {
	ws.cancel()
	return ws.tryClose()
}

func (ws *Websocket) tryClose() error {
	ws.connMu.Lock()
	defer ws.connMu.Unlock()

	var err error
	if ws.conn != nil {
		writeErr := ws.conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(time.Second))
		closeErr := ws.conn.Close()
		err = errors.Join(writeErr, closeErr)
	}

	return err
}
