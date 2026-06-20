package server

import (
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/prabhatlabs/go-tunnel/internal/protocol"
)

const PROXY_TIMEOUT = 30 * time.Second

func uniqueID() string {
	return uuid.New().String()
}

func (s *Server) handleProxy(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	t := s.tunnel
	s.mu.Unlock()
	if t == nil {
		http.Error(w, "No tunnel", http.StatusBadGateway)
		return
	}

	id := uniqueID()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusInternalServerError)
		return
	}

	reqMessage := &protocol.RequestMessage{
		ID:      id,
		Type:    "request",
		Method:  r.Method,
		Path:    r.URL.Path,
		Headers: make(map[string]string),
		Body:    body,
	}

	// adding headers to request message
	for k, v := range r.Header {
		reqMessage.Headers[k] = v[0]
	}

	// sending request message to tunnel
	ch, err := t.SendRequest(reqMessage)
	if err != nil {
		http.Error(w, "Failed to send request", http.StatusInternalServerError)
		return
	}

	select {
	case resp := <-ch:
		for k, v := range resp.Headers {
			w.Header().Set(k, v)
		}
		w.WriteHeader(resp.Status)
		w.Write(resp.Body)
	case <-time.After(PROXY_TIMEOUT):
		http.Error(w, "Timeout", http.StatusGatewayTimeout)
	}
}
