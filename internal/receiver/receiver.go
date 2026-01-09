package receiver

import (
	"io"
	"net/http"
)

const maxBodySize = 1 << 20 // 1MB

type Message struct {
	Path string
	Data []byte
}

type Receiver struct {
	out chan<- Message
}

func New(out chan<- Message) *Receiver {
	return &Receiver{out: out}
}

func (r *Receiver) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	if req.ContentLength > maxBodySize {
		http.Error(w, "Request Entity Too Large", http.StatusRequestEntityTooLarge)
		return
	}

	body, err := io.ReadAll(io.LimitReader(req.Body, maxBodySize+1))
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	if len(body) > maxBodySize {
		http.Error(w, "Request Entity Too Large", http.StatusRequestEntityTooLarge)
		return
	}

	r.out <- Message{Path: req.URL.Path, Data: body}

	w.WriteHeader(http.StatusOK)
}
