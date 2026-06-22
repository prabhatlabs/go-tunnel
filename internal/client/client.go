package client

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/prabhatlabs/go-tunnel/internal/protocol"
)

type Client struct {
	ServerUrl string
	Port      int
	Secret    string
}

func New(serverUrl string, port int, secret string) *Client {
	return &Client{
		ServerUrl: serverUrl,
		Port:      port,
		Secret:    secret,
	}
}

func (c *Client) connect() (*websocket.Conn, error) {
	serverUrl := "ws://" + c.ServerUrl + "/__tunnel__"

	headers := make(http.Header)
	headers.Add("X-Tunnel-Secret", c.Secret)

	conn, resp, err := websocket.DefaultDialer.Dial(serverUrl, headers)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusSwitchingProtocols {
		return nil, err
	}
	return conn, nil
}

func (c *Client) readLoop(conn *websocket.Conn, done chan struct{}, writeCh chan []byte) {
	for {
		select {
		case <-done:
			return
		default:
		}
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}

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

		writeCh <- data
	}
}

func (c *Client) writeLoop(conn *websocket.Conn, done chan struct{}, writeCh chan []byte) {
	for {
		select {
		case <-done:
			return
		case msg := <-writeCh:
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		}
	}
}

func (c *Client) runSession(conn *websocket.Conn) {
	done := make(chan struct{})
	writeCh := make(chan []byte)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		c.readLoop(conn, done, writeCh)
	}()
	go func() {
		defer wg.Done()
		c.writeLoop(conn, done, writeCh)
	}()

	wg.Wait()
	conn.Close()
}

func (c *Client) Run() {
	backoff := 1 * time.Second

	for {
		conn, err := c.connect()
		if err != nil {
			log.Printf("connect failed: %v, retrying in %v", err, backoff)
			time.Sleep(backoff)
			backoff *= 2
			if backoff > 30*time.Second {
				backoff = 30 * time.Second
			}
			continue
		}

		backoff = 1 * time.Second
		log.Printf("connected to tunnel server")

		c.runSession(conn)
	}
}
