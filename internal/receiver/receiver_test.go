package receiver

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestServeHTTP_MethodNotAllowed(t *testing.T) {
	ch := make(chan Message, 1)
	r := New(ch)

	methods := []string{"GET", "PUT", "DELETE", "PATCH"}
	for _, method := range methods {
		req := httptest.NewRequest(method, "/webhook", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s: got %d, want %d", method, rec.Code, http.StatusMethodNotAllowed)
		}
	}
}

func TestServeHTTP_Success(t *testing.T) {
	ch := make(chan Message, 1)
	r := New(ch)

	body := []byte(`{"event": "push"}`)
	req := httptest.NewRequest("POST", "/webhook/test", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got %d, want %d", rec.Code, http.StatusOK)
	}

	select {
	case msg := <-ch:
		if msg.Path != "/webhook/test" {
			t.Errorf("path: got %q, want %q", msg.Path, "/webhook/test")
		}
		if !bytes.Equal(msg.Data, body) {
			t.Errorf("data: got %q, want %q", msg.Data, body)
		}
	case <-time.After(time.Second):
		t.Error("timeout waiting for message")
	}
}

func TestServeHTTP_BodyWithNewlines(t *testing.T) {
	ch := make(chan Message, 1)
	r := New(ch)

	body := []byte("line1\nline2\nline3")
	req := httptest.NewRequest("POST", "/test", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got %d, want %d", rec.Code, http.StatusOK)
	}

	select {
	case msg := <-ch:
		if !bytes.Equal(msg.Data, body) {
			t.Errorf("data: got %q, want %q", msg.Data, body)
		}
	case <-time.After(time.Second):
		t.Error("timeout waiting for message")
	}
}

func TestServeHTTP_TooLarge(t *testing.T) {
	ch := make(chan Message, 1)
	r := New(ch)

	body := strings.Repeat("x", maxBodySize+1)
	req := httptest.NewRequest("POST", "/test", strings.NewReader(body))
	req.ContentLength = int64(len(body))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("got %d, want %d", rec.Code, http.StatusRequestEntityTooLarge)
	}

	select {
	case <-ch:
		t.Error("should not receive message for oversized body")
	default:
	}
}
