package router

import (
	"os"
	"strings"

	"github.com/aster-void/webhooker/internal/receiver"
)

type cmd struct {
	register   bool
	unregister bool
	path       string
	ch         chan<- []byte
}

type Router struct {
	in     <-chan receiver.Message
	cmd    chan cmd
	routes map[string]chan<- []byte // path -> writer channel
}

func New(in <-chan receiver.Message) *Router {
	return &Router{
		in:     in,
		cmd:    make(chan cmd, 100),
		routes: make(map[string]chan<- []byte),
	}
}

func (r *Router) LoadRoutes(writers ...chan<- []byte) {
	env := os.Getenv("WEBHOOKER_ROUTES")
	if env == "" {
		return
	}

	for _, pair := range strings.Split(env, ",") {
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) == 2 {
			secret := strings.TrimSpace(parts[0])
			if secret != "" {
				r.routes["/"+secret] = fanout(writers)
			}
		}
	}
}

func (r *Router) Register(path string, ch chan<- []byte) {
	r.cmd <- cmd{register: true, path: path, ch: ch}
}

func (r *Router) Unregister(path string) {
	r.cmd <- cmd{unregister: true, path: path}
}

func (r *Router) Run() {
	for {
		select {
		case msg, ok := <-r.in:
			if !ok {
				return
			}
			r.route(msg)
		case c := <-r.cmd:
			if c.register {
				r.routes[c.path] = c.ch
			}
			if c.unregister {
				if ch, ok := r.routes[c.path]; ok {
					close(ch)
					delete(r.routes, c.path)
				}
			}
		}
	}
}

func (r *Router) route(msg receiver.Message) {
	ch, ok := r.routes[msg.Path]
	if !ok {
		return // no matching route, silently drop
	}
	ch <- msg.Data
}

// fanout creates a channel that writes to multiple channels
func fanout(chs []chan<- []byte) chan<- []byte {
	out := make(chan []byte, 100)
	go func() {
		for data := range out {
			for _, ch := range chs {
				ch <- data
			}
		}
	}()
	return out
}
