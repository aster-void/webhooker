package receiver

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
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

	timestamp := time.Now().UTC().Format(time.RFC3339Nano)
	escaped := escape(body)
	line := []byte(fmt.Sprintf("%s %s %s %s\n", timestamp, req.Method, req.URL.Path, escaped))

	r.out <- Message{Path: req.URL.Path, Data: line}

	w.WriteHeader(http.StatusOK)
}

func escape(data []byte) string {
	var b strings.Builder
	b.Grow(len(data))

	for _, c := range data {
		switch {
		case c == '\\':
			b.WriteString(`\\`)
		case c == '\n':
			b.WriteString(`\n`)
		case c == '\r':
			b.WriteString(`\r`)
		case c == '\t':
			b.WriteString(`\t`)
		case c < 0x20 || c > 0x7e:
			fmt.Fprintf(&b, `\x%02x`, c)
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}
