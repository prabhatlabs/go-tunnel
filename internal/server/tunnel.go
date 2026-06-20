package server

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/prabhatlabs/go-tunnel/internal/protocol"
)

type Tunnel struct {
	conn    *websocket.Conn
	pending sync.Map
	writeCh chan []byte
	done    chan struct{}
}

func NewTunnel(c *websocket.Conn) *Tunnel {
	return &Tunnel{
		conn:    c,
		pending: sync.Map{},
		writeCh: make(chan []byte),
		done:    make(chan struct{}),
	}
}

func (t *Tunnel) Close() {
	select {
	case <-t.done:
		return
	default:
		close(t.done)
	}

	t.conn.Close()

	close(t.writeCh)
	t.pending.Range(func(k, value any) bool {
		ch := value
		resp := protocol.ResponseMessage{
			Status:  http.StatusBadGateway,
			Headers: nil,
			Body:    nil,
		}
		ch.(chan *protocol.ResponseMessage) <- &resp
		t.pending.Delete(k)
		return true
	})

	// drain remaining(if there are any) writeCh
	for {
		select {
		case <-t.writeCh:
		default:
			return
		}
	}
}

func (t *Tunnel) ReadLoop() {
	for {
		select {
		case <-t.done:
			return
		default:
			_, msg, err := t.conn.ReadMessage()
			if err != nil {
				return
			}

			var resp protocol.ResponseMessage
			if err := resp.UnmarshalJSON(msg); err != nil {
				continue
			}
			ch, ok := t.pending.Load(resp.ID)
			if !ok {
				continue
			}
			ch.(chan *protocol.ResponseMessage) <- &resp
			t.pending.Delete(resp.ID)
		}
	}
}

func (t *Tunnel) WriteLoop(msgCh <-chan []byte) {
	for {
		select {
		case <-t.done:
			return
		case msg := <-msgCh:
			if err := t.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		}
	}
}

func (t *Tunnel) SendRequest(req *protocol.RequestMessage) (<-chan *protocol.ResponseMessage, error) {
	data, err := req.MarshalJSON()
	if err != nil {
		return nil, err
	}

	ch := make(chan *protocol.ResponseMessage, 1)
	t.pending.Store(req.ID, ch)

	// queue the request to be sent
	t.writeCh <- data
	return ch, nil
}
