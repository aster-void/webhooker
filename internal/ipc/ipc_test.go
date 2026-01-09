package ipc

import (
	"bufio"
	"encoding/json"
	"net"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestServer_RegisterAndWebhook(t *testing.T) {
	dir := t.TempDir()
	sockPath := filepath.Join(dir, "test.sock")

	var registeredPath string
	var registeredCh chan<- []byte

	register := func(path string, ch chan<- []byte) {
		registeredPath = path
		registeredCh = ch
	}
	unregister := func(path string) {}

	srv, err := NewServer(sockPath, "https://example.com", register, unregister)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	defer srv.Close()
	go srv.Run()

	// Connect client
	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	encoder := json.NewEncoder(conn)
	reader := bufio.NewReader(conn)

	// Send register request
	if err := encoder.Encode(Request{Type: "register"}); err != nil {
		t.Fatalf("send register: %v", err)
	}

	// Read registered response
	line, err := reader.ReadBytes('\n')
	if err != nil {
		t.Fatalf("read response: %v", err)
	}

	var resp Response
	if err := json.Unmarshal(line, &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if resp.Type != "registered" {
		t.Errorf("type: got %q, want %q", resp.Type, "registered")
	}
	if !strings.HasPrefix(resp.Path, "/tmp-") {
		t.Errorf("path: got %q, want prefix /tmp-", resp.Path)
	}
	if !strings.HasPrefix(resp.URL, "https://example.com/tmp-") {
		t.Errorf("url: got %q, want prefix https://example.com/tmp-", resp.URL)
	}

	// Verify registration callback was called
	if registeredPath == "" {
		t.Error("register callback not called")
	}

	// Send webhook data
	if registeredCh != nil {
		registeredCh <- []byte("test webhook data")
	}

	// Read webhook response
	conn.SetReadDeadline(time.Now().Add(time.Second))
	line, err = reader.ReadBytes('\n')
	if err != nil {
		t.Fatalf("read webhook: %v", err)
	}

	if err := json.Unmarshal(line, &resp); err != nil {
		t.Fatalf("unmarshal webhook: %v", err)
	}

	if resp.Type != "webhook" {
		t.Errorf("type: got %q, want %q", resp.Type, "webhook")
	}
	if resp.Data != "test webhook data" {
		t.Errorf("data: got %q, want %q", resp.Data, "test webhook data")
	}
}

func TestServer_WebhookWithNewlines(t *testing.T) {
	dir := t.TempDir()
	sockPath := filepath.Join(dir, "test.sock")

	var registeredCh chan<- []byte

	register := func(path string, ch chan<- []byte) {
		registeredCh = ch
	}
	unregister := func(path string) {}

	srv, err := NewServer(sockPath, "", register, unregister)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	defer srv.Close()
	go srv.Run()

	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	encoder := json.NewEncoder(conn)
	reader := bufio.NewReader(conn)

	encoder.Encode(Request{Type: "register"})
	reader.ReadBytes('\n') // skip registered response

	// Send data with newlines
	data := "line1\nline2\nline3"
	if registeredCh != nil {
		registeredCh <- []byte(data)
	}

	conn.SetReadDeadline(time.Now().Add(time.Second))
	line, err := reader.ReadBytes('\n')
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	var resp Response
	if err := json.Unmarshal(line, &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if resp.Data != data {
		t.Errorf("data: got %q, want %q", resp.Data, data)
	}
}

func TestServer_InvalidRequest(t *testing.T) {
	dir := t.TempDir()
	sockPath := filepath.Join(dir, "test.sock")

	srv, err := NewServer(sockPath, "", func(string, chan<- []byte) {}, func(string) {})
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	defer srv.Close()
	go srv.Run()

	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	// Send invalid request type
	encoder := json.NewEncoder(conn)
	reader := bufio.NewReader(conn)

	encoder.Encode(Request{Type: "invalid"})

	conn.SetReadDeadline(time.Now().Add(time.Second))
	line, err := reader.ReadBytes('\n')
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	var resp Response
	json.Unmarshal(line, &resp)

	if resp.Type != "error" {
		t.Errorf("type: got %q, want %q", resp.Type, "error")
	}
}
