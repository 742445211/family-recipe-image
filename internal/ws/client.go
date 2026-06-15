package ws

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"recipe-image/internal/config"
	"recipe-image/internal/dispatcher"
	"recipe-image/internal/protocol"

	"github.com/gorilla/websocket"
)

type Client struct {
	cfg        config.ServerConfig
	dispatcher *dispatcher.Dispatcher
	conn       *websocket.Conn
	mu         sync.Mutex
}

func NewClient(cfg config.ServerConfig, disp *dispatcher.Dispatcher) *Client {
	return &Client{cfg: cfg, dispatcher: disp}
}

func (c *Client) Run(ctx context.Context) error {
	backoff := time.Duration(c.cfg.ReconnectSec) * time.Second
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err := c.connectOnce(ctx); err != nil {
			log.Printf("[ws] connection error: %v, retry in %s", err, backoff)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}
		if backoff < 60*time.Second {
			backoff += time.Duration(c.cfg.ReconnectSec) * time.Second
		}
	}
}

func (c *Client) connectOnce(ctx context.Context) error {
	u, err := url.Parse(c.cfg.WsURL)
	if err != nil {
		return err
	}
	if u.Scheme != "wss" {
		return fmt.Errorf("ws_url must use wss://")
	}
	q := u.Query()
	q.Set("token", c.cfg.Token)
	if c.cfg.WorkerID != "" {
		q.Set("worker_id", c.cfg.WorkerID)
	}
	u.RawQuery = q.Encode()

	dialer := websocket.Dialer{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: u.Hostname(),
		},
	}
	conn, _, err := dialer.DialContext(ctx, u.String(), nil)
	if err != nil {
		return err
	}
	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()
	defer func() {
		c.mu.Lock()
		c.conn = nil
		c.mu.Unlock()
		_ = conn.Close()
	}()

	log.Printf("[ws] connected to %s", c.cfg.WsURL)

	pingCtx, pingCancel := context.WithCancel(ctx)
	defer pingCancel()
	go c.pingLoop(pingCtx, conn)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		_, data, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		if err := c.handleMessage(ctx, data); err != nil {
			log.Printf("[ws] handle message: %v", err)
		}
	}
}

func (c *Client) pingLoop(ctx context.Context, conn *websocket.Conn) {
	ticker := time.NewTicker(time.Duration(c.cfg.PingIntervalSec) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.mu.Lock()
			err := conn.WriteJSON(protocol.PingMessage{Type: protocol.TypePing})
			c.mu.Unlock()
			if err != nil {
				return
			}
		}
	}
}

func (c *Client) handleMessage(ctx context.Context, data []byte) error {
	var base struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &base); err != nil {
		return err
	}
	switch base.Type {
	case protocol.TypeTask:
		if c.dispatcher == nil {
			return fmt.Errorf("dispatcher not ready")
		}
		task, err := protocol.ParseTask(data)
		if err != nil {
			return err
		}
		c.dispatcher.Submit(ctx, task)
	case protocol.TypeRegistered:
		var msg protocol.RegisteredMessage
		_ = json.Unmarshal(data, &msg)
		log.Printf("[ws] registered worker_id=%s", msg.WorkerID)
	case protocol.TypePong:
		return nil
	default:
		if strings.TrimSpace(base.Type) != "" {
			log.Printf("[ws] ignored message type=%s", base.Type)
		}
	}
	return nil
}

func (c *Client) SendResult(msg *protocol.TaskResultMessage) {
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()
	if conn == nil {
		log.Printf("[ws] drop result (offline) task=%s", msg.TaskID)
		return
	}
	c.mu.Lock()
	err := conn.WriteJSON(msg)
	c.mu.Unlock()
	if err != nil {
		log.Printf("[ws] send result failed: %v", err)
	}
}
