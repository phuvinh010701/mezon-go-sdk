// Package websocket provides the WebSocket client for the Mezon real-time event system.
//
// Usage:
//
//	conn, err := websocket.New(token, websocket.WithBaseURL("wss://api.mezon.ai"))
//	if err != nil { ... }
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	if err := conn.Connect(ctx); err != nil { ... }
//	defer conn.Close()
//
//	conn.OnEvent(models.EventChannelMessage, func(payload json.RawMessage) {
//	    var msg models.ChannelMessage
//	    json.Unmarshal(payload, &msg)
//	    // handle message
//	})
//
//	if err := conn.Send(ctx, envelope); err != nil { ... }
package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/phuvinh010701/mezon-go-sdk/models"
)

const (
	defaultWSBaseURL       = "wss://api.mezon.ai"
	defaultPingInterval    = 10 * time.Second
	defaultWriteTimeout    = 5 * time.Second
	defaultPongTimeout     = 15 * time.Second
	defaultReconnectDelay  = 1500 * time.Millisecond
	defaultMaxReconnectDelay = 15 * time.Second
	defaultMaxReconnects   = 0 // 0 = unlimited
)

// HandlerFunc is called when an event of the registered type is received.
type HandlerFunc func(payload json.RawMessage)

// Envelope is the wire format for all WebSocket messages.
// Fields map to the Mezon protobuf Envelope as JSON.
type Envelope struct {
	// CID is the client-assigned command ID used to correlate responses.
	CID string `json:"cid,omitempty"`
	// Type is the event type string (maps to models.Event).
	Type string `json:"type,omitempty"`
	// Payload holds the event-specific JSON data.
	Payload json.RawMessage `json:"payload,omitempty"`
}

// Option configures a Conn.
type Option func(*Conn)

// WithBaseURL overrides the default WebSocket base URL.
func WithBaseURL(url string) Option {
	return func(c *Conn) { c.baseURL = url }
}

// WithLogger sets a custom structured logger.
func WithLogger(l *slog.Logger) Option {
	return func(c *Conn) { c.logger = l }
}

// WithPingInterval sets how often keep-alive pings are sent.
func WithPingInterval(d time.Duration) Option {
	return func(c *Conn) { c.pingInterval = d }
}

// WithAutoReconnect enables automatic reconnection on disconnect.
// maxAttempts of 0 means unlimited retries.
func WithAutoReconnect(maxAttempts int) Option {
	return func(c *Conn) {
		c.autoReconnect = true
		c.maxReconnects = maxAttempts
	}
}

// Conn manages a WebSocket connection to the Mezon real-time API.
// It is safe for concurrent use after Connect is called.
type Conn struct {
	token        string
	baseURL      string
	logger       *slog.Logger
	pingInterval time.Duration

	autoReconnect bool
	maxReconnects int

	mu       sync.RWMutex
	ws       *websocket.Conn
	handlers map[models.Event][]HandlerFunc

	pendingMu sync.Mutex
	pending   map[string]chan json.RawMessage

	cidCounter atomic.Uint64

	// writeMu serialises writes; gorilla/websocket does not support concurrent writes.
	writeMu sync.Mutex
}

// New creates a new Conn. token must be a valid Mezon JWT bearer token.
func New(token string, opts ...Option) (*Conn, error) {
	if token == "" {
		return nil, fmt.Errorf("websocket: token must not be empty")
	}
	c := &Conn{
		token:        token,
		baseURL:      defaultWSBaseURL,
		logger:       slog.Default(),
		pingInterval: defaultPingInterval,
		handlers:     make(map[models.Event][]HandlerFunc),
		pending:      make(map[string]chan json.RawMessage),
	}
	for _, o := range opts {
		o(c)
	}
	return c, nil
}

// Connect dials the Mezon WebSocket endpoint and starts the read/ping loops.
// The loops run until ctx is cancelled or the connection is closed.
func (c *Conn) Connect(ctx context.Context) error {
	return c.connectOnce(ctx)
}

func (c *Conn) connectOnce(ctx context.Context) error {
	url := fmt.Sprintf("%s/ws?lang=en&status=true&token=%s&format=json", c.baseURL, c.token)

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}
	header := http.Header{}
	header.Set("Authorization", "Bearer "+c.token)

	ws, _, err := dialer.DialContext(ctx, url, header)
	if err != nil {
		return fmt.Errorf("websocket: dial %s: %w", c.baseURL, err)
	}

	ws.SetPongHandler(func(string) error {
		return ws.SetReadDeadline(time.Now().Add(defaultPongTimeout))
	})

	c.mu.Lock()
	c.ws = ws
	c.mu.Unlock()

	c.logger.Info("websocket connected", "url", c.baseURL)

	// Start background loops.
	go c.readLoop(ctx)
	go c.pingLoop(ctx)

	return nil
}

// Close sends a close frame and tears down the connection.
func (c *Conn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ws == nil {
		return nil
	}

	c.writeMu.Lock()
	_ = c.ws.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
	)
	c.writeMu.Unlock()

	err := c.ws.Close()
	c.ws = nil
	return err
}

// IsOpen reports whether the connection is active.
func (c *Conn) IsOpen() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ws != nil
}

// OnEvent registers handler to be called for every event of the given type.
// Handlers are called in a new goroutine and must not block for extended periods.
// Multiple handlers for the same event are all invoked.
func (c *Conn) OnEvent(event models.Event, handler HandlerFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handlers[event] = append(c.handlers[event], handler)
}

// OffEvent removes all handlers registered for event.
func (c *Conn) OffEvent(event models.Event) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.handlers, event)
}

// Send writes an Envelope to the WebSocket. It is safe for concurrent use.
func (c *Conn) Send(ctx context.Context, env Envelope) error {
	data, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("websocket: marshal envelope: %w", err)
	}

	c.mu.RLock()
	ws := c.ws
	c.mu.RUnlock()

	if ws == nil {
		return fmt.Errorf("websocket: connection not open")
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	if err := ws.SetWriteDeadline(time.Now().Add(defaultWriteTimeout)); err != nil {
		return fmt.Errorf("websocket: set write deadline: %w", err)
	}
	if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
		return fmt.Errorf("websocket: write: %w", err)
	}
	return nil
}

// Request sends an Envelope and waits for a response with matching CID.
// It returns the raw response payload or an error if ctx is cancelled or
// the timeout is exceeded.
func (c *Conn) Request(ctx context.Context, env Envelope) (json.RawMessage, error) {
	if env.CID == "" {
		env.CID = c.nextCID()
	}

	ch := make(chan json.RawMessage, 1)

	c.pendingMu.Lock()
	c.pending[env.CID] = ch
	c.pendingMu.Unlock()

	defer func() {
		c.pendingMu.Lock()
		delete(c.pending, env.CID)
		c.pendingMu.Unlock()
	}()

	if err := c.Send(ctx, env); err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("websocket request: %w", ctx.Err())
	case raw := <-ch:
		return raw, nil
	}
}

// nextCID generates a unique command ID string.
func (c *Conn) nextCID() string {
	return fmt.Sprintf("%d", c.cidCounter.Add(1))
}

// readLoop continuously reads messages and dispatches them.
func (c *Conn) readLoop(ctx context.Context) {
	attempts := 0
	for {
		if err := ctx.Err(); err != nil {
			return
		}

		c.mu.RLock()
		ws := c.ws
		c.mu.RUnlock()

		if ws == nil {
			return
		}

		_ = ws.SetReadDeadline(time.Now().Add(defaultPongTimeout))
		_, data, err := ws.ReadMessage()
		if err != nil {
			c.logger.Warn("websocket read error", "error", err)

			// Cancel all pending requests.
			c.pendingMu.Lock()
			for _, ch := range c.pending {
				close(ch)
			}
			c.pending = make(map[string]chan json.RawMessage)
			c.pendingMu.Unlock()

			if !c.autoReconnect {
				return
			}
			if c.maxReconnects > 0 && attempts >= c.maxReconnects {
				c.logger.Error("websocket max reconnect attempts reached")
				return
			}

			delay := backoffDelay(attempts)
			c.logger.Info("websocket reconnecting", "delay", delay, "attempt", attempts+1)

			select {
			case <-ctx.Done():
				return
			case <-time.After(delay):
			}

			if reconnErr := c.connectOnce(ctx); reconnErr != nil {
				c.logger.Error("websocket reconnect failed", "error", reconnErr)
				attempts++
				continue
			}
			attempts = 0
			continue
		}

		var env Envelope
		if jsonErr := json.Unmarshal(data, &env); jsonErr != nil {
			c.logger.Debug("websocket: failed to unmarshal envelope", "error", jsonErr)
			continue
		}

		// If this is a response to a pending request, route it.
		if env.CID != "" {
			c.pendingMu.Lock()
			ch, ok := c.pending[env.CID]
			c.pendingMu.Unlock()
			if ok {
				select {
				case ch <- env.Payload:
				default:
				}
				continue
			}
		}

		// Otherwise dispatch to registered event handlers.
		if env.Type != "" {
			c.dispatch(models.Event(env.Type), env.Payload)
		}
	}
}

// pingLoop sends periodic pings to keep the connection alive.
func (c *Conn) pingLoop(ctx context.Context) {
	ticker := time.NewTicker(c.pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.mu.RLock()
			ws := c.ws
			c.mu.RUnlock()

			if ws == nil {
				return
			}

			c.writeMu.Lock()
			_ = ws.SetWriteDeadline(time.Now().Add(defaultWriteTimeout))
			if err := ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.logger.Warn("websocket: ping failed", "error", err)
				c.writeMu.Unlock()
				return
			}
			c.writeMu.Unlock()
		}
	}
}

// dispatch calls all registered handlers for event in separate goroutines.
func (c *Conn) dispatch(event models.Event, payload json.RawMessage) {
	c.mu.RLock()
	handlers := make([]HandlerFunc, len(c.handlers[event]))
	copy(handlers, c.handlers[event])
	c.mu.RUnlock()

	for _, h := range handlers {
		h := h
		go func() {
			defer func() {
				if r := recover(); r != nil {
					c.logger.Error("websocket: handler panic", "event", event, "panic", r)
				}
			}()
			h(payload)
		}()
	}
}

// backoffDelay returns the reconnection wait time for the given attempt index.
func backoffDelay(attempt int) time.Duration {
	d := defaultReconnectDelay * time.Duration(1<<uint(attempt))
	if d > defaultMaxReconnectDelay {
		d = defaultMaxReconnectDelay
	}
	return d
}
