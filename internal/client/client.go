package client

import (
	"errors"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/prabhatlabs/go-tunnel/internal/protocol"
)

type Client struct {
	ServerUrl string
	Port      int
	Secret    string

	mu      *sync.Mutex
	conn    *websocket.Conn
	writeCh chan []byte
	done    chan struct{}
}

func New(serverUrl string, port int, secret string) *Client {
	return &Client{
		ServerUrl: serverUrl,
		Port:      port,
		Secret:    secret,
		mu:        &sync.Mutex{},
		writeCh:   make(chan []byte),
		done:      make(chan struct{}),
	}
}

func (c *Client) connect() (*websocket.Conn, error) {
	serverUrl := "wss://" + c.ServerUrl + "/__tunnel__"

	headers := make(http.Header)
	headers.Add("X-Tunnel-Secret", c.Secret)

	conn, resp, err := websocket.DefaultDialer.Dial(serverUrl, headers)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusSwitchingProtocols {
		return nil, errors.New("Unexpected status code from tunnel server: " + resp.Status)
	}
	return conn, nil
}

func (c *Client) close() {
	select {
	case <-c.done:
		return
	default:
		close(c.done)
	}

	if c.conn == nil {
		return
	}

	if err := c.conn.Close(); err != nil {
		return
	}

	c.mu.Lock()
	c.conn = nil
	c.mu.Unlock()
}

func (c *Client) ReadLoop(onClose func()) {
	defer func() {
		if onClose != nil {
			onClose()
		}
	}()

	for {
		select {
		case <-c.done:
			return
		default:
		}
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
		// forwarding request to local service
		var req protocol.RequestMessage
		if err := req.UnmarshalJSON(data); err != nil {
			continue
		}

		resp, err := httpRequestBuilderAndDoer(c.Port, &req)
		if err != nil {
			continue
		}

		data, err = resp.MarshalJSON()
		if err != nil {
			continue
		}

		c.writeCh <- data
	}
}

func (c *Client) WriteLoop(msgCh chan []byte) {
	for {
		select {
		case <-c.done:
			return
		case msg := <-msgCh:
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		}
	}
}

func (c *Client) Run() error {
	conn, err := c.connect()
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()

	go c.ReadLoop(c.close)
	go c.WriteLoop(c.writeCh)
	return nil
}
