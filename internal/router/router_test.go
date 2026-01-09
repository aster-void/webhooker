package router

import (
	"testing"
	"time"

	"github.com/aster-void/webhooker/internal/receiver"
)

func TestRouter_RegisterAndRoute(t *testing.T) {
	in := make(chan receiver.Message, 10)
	r := New(in)
	go r.Run()
	defer close(in)

	out := make(chan []byte, 10)
	r.Register("/test", out)

	// Allow registration to process
	time.Sleep(10 * time.Millisecond)

	in <- receiver.Message{Path: "/test", Data: []byte("hello")}

	select {
	case data := <-out:
		if string(data) != "hello" {
			t.Errorf("got %q, want %q", data, "hello")
		}
	case <-time.After(time.Second):
		t.Error("timeout waiting for routed message")
	}
}

func TestRouter_UnregisteredPath(t *testing.T) {
	in := make(chan receiver.Message, 10)
	r := New(in)
	go r.Run()
	defer close(in)

	out := make(chan []byte, 10)
	r.Register("/registered", out)

	time.Sleep(10 * time.Millisecond)

	in <- receiver.Message{Path: "/unregistered", Data: []byte("ignored")}

	select {
	case <-out:
		t.Error("should not receive message for unregistered path")
	case <-time.After(50 * time.Millisecond):
		// Expected: message silently ignored
	}
}

func TestRouter_Unregister(t *testing.T) {
	in := make(chan receiver.Message, 10)
	r := New(in)
	go r.Run()
	defer close(in)

	out := make(chan []byte, 10)
	r.Register("/test", out)

	time.Sleep(10 * time.Millisecond)

	r.Unregister("/test")

	time.Sleep(10 * time.Millisecond)

	// Channel should be closed
	select {
	case _, ok := <-out:
		if ok {
			t.Error("channel should be closed after unregister")
		}
	case <-time.After(time.Second):
		t.Error("timeout waiting for channel close")
	}
}

func TestRouter_MultipleRoutes(t *testing.T) {
	in := make(chan receiver.Message, 10)
	r := New(in)
	go r.Run()
	defer close(in)

	out1 := make(chan []byte, 10)
	out2 := make(chan []byte, 10)
	r.Register("/path1", out1)
	r.Register("/path2", out2)

	time.Sleep(10 * time.Millisecond)

	in <- receiver.Message{Path: "/path1", Data: []byte("msg1")}
	in <- receiver.Message{Path: "/path2", Data: []byte("msg2")}

	select {
	case data := <-out1:
		if string(data) != "msg1" {
			t.Errorf("out1: got %q, want %q", data, "msg1")
		}
	case <-time.After(time.Second):
		t.Error("timeout waiting for out1")
	}

	select {
	case data := <-out2:
		if string(data) != "msg2" {
			t.Errorf("out2: got %q, want %q", data, "msg2")
		}
	case <-time.After(time.Second):
		t.Error("timeout waiting for out2")
	}
}
