package ipc

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
)

const defaultSocket = "/var/lib/webhooker/webhooker.sock"

func RunClient(socketPath string) error {
	if socketPath == "" {
		socketPath = defaultSocket
	}
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return fmt.Errorf("connect to daemon: %w (is daemon running?)", err)
	}
	defer conn.Close()

	encoder := json.NewEncoder(conn)
	reader := bufio.NewReader(conn)

	// Send register request
	if err := encoder.Encode(Request{Type: "register"}); err != nil {
		return fmt.Errorf("send register: %w", err)
	}

	// Read responses
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return nil // connection closed
		}

		var resp Response
		if err := json.Unmarshal(line, &resp); err != nil {
			continue
		}

		switch resp.Type {
		case "registered":
			fmt.Fprintf(os.Stderr, "listening on %s\n", resp.Path)
		case "webhook":
			fmt.Print(resp.Data)
		case "error":
			return fmt.Errorf("server error: %s", resp.Data)
		}
	}
}
