package ipc

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
)

const DefaultSocket = "/var/lib/webhooker/webhooker.sock"

type RegisterFunc func(path string, ch chan<- []byte)
type UnregisterFunc func(path string)

type Server struct {
	listener   net.Listener
	path       string
	domain     string
	register   RegisterFunc
	unregister UnregisterFunc
}

func NewServer(socketPath, domain string, register RegisterFunc, unregister UnregisterFunc) (*Server, error) {
	if socketPath == "" {
		socketPath = DefaultSocket
	}

	dir := filepath.Dir(socketPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, err
	}

	return &Server{
		listener:   listener,
		path:       socketPath,
		domain:     domain,
		register:   register,
		unregister: unregister,
	}, nil
}

func (s *Server) Run() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return
		}
		go s.handle(conn)
	}
}

func (s *Server) Close() error {
	s.listener.Close()
	os.Remove(s.path)
	return nil
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	encoder := json.NewEncoder(conn)

	// Read register request
	line, err := reader.ReadBytes('\n')
	if err != nil {
		return
	}

	var req Request
	if err := json.Unmarshal(line, &req); err != nil {
		encoder.Encode(Response{Type: "error", Data: "invalid json"})
		return
	}

	if req.Type != "register" {
		encoder.Encode(Response{Type: "error", Data: "expected register"})
		return
	}

	// Generate temp path
	path := "/tmp-" + randomHex(8)
	ch := make(chan []byte, 100)

	s.register(path, ch)
	defer s.unregister(path)

	// Send registered response
	resp := Response{Type: "registered", Path: path}
	if s.domain != "" {
		resp.URL = s.domain + path
	}
	encoder.Encode(resp)

	// Stream webhooks to client
	for data := range ch {
		if err := encoder.Encode(Response{Type: "webhook", Data: string(data)}); err != nil {
			return
		}
	}
}

func randomHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}
