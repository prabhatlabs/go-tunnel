package server

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
)

type Server struct {
	Port     int
	Secret   string
	mu       sync.Mutex
	mux      *http.ServeMux
	upgrader *websocket.Upgrader
	tunnel   *Tunnel
}

func New(port int, secret string) *Server {
	svr := &Server{
		Port:   port,
		Secret: secret,

		mux: http.NewServeMux(),
		upgrader: &websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}

	svr.mux.HandleFunc("/__tunnel__", svr.handleTunnel)
	svr.mux.HandleFunc("/", svr.handleProxy)
	return svr
}

func (s *Server) handleTunnel(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	if s.tunnel != nil {
		s.mu.Unlock()
		http.Error(w, "Already connected", http.StatusConflict)
		return
	}

	tS := r.Header.Get("X-Tunnel-Secret")
	if tS != s.Secret {
		s.mu.Unlock()
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ws, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.mu.Unlock()
		http.Error(w, "Upgrade failed", http.StatusInternalServerError)
		return
	}

	s.tunnel = NewTunnel(ws)
	s.mu.Unlock()
	defer s.closeTunnel()
	go s.tunnel.ReadLoop()
	go s.tunnel.WriteLoop(s.tunnel.writeCh)
}

func (s *Server) closeTunnel() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.tunnel == nil {
		return
	}
	s.tunnel.Close()
	s.tunnel = nil
}

func (s *Server) Run() error {
	p := s.Port
	if p == 0 {
		p = 8080
	}
	return http.ListenAndServe(":"+strconv.Itoa(p), s.mux)
}
