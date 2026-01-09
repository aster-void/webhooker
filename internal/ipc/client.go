package ipc

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
)

func RunClient() error {
	conn, err := dialSocket()
	if err != nil {
		return err
	}
	defer conn.Close()

	encoder := json.NewEncoder(conn)
	reader := bufio.NewReader(conn)

	if err := encoder.Encode(Request{Type: "register"}); err != nil {
		return fmt.Errorf("send register: %w", err)
	}

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return nil
		}

		var resp Response
		if err := json.Unmarshal(line, &resp); err != nil {
			continue
		}

		switch resp.Type {
		case "registered":
			addr := resp.URL
			if addr == "" {
				addr = resp.Path
			}
			fmt.Fprintf(os.Stderr, "listening on %s\n", addr)
		case "webhook":
			fmt.Print(resp.Data)
		case "error":
			return fmt.Errorf("server error: %s", resp.Data)
		}
	}
}

func dialSocket() (net.Conn, error) {
	paths := []string{"/run/webhooker/webhooker.sock"}
	if xdg := os.Getenv("XDG_RUNTIME_DIR"); xdg != "" {
		paths = append(paths, xdg+"/webhooker/webhooker.sock")
	}

	for _, p := range paths {
		conn, err := net.Dial("unix", p)
		if err == nil {
			return conn, nil
		}
	}
	return nil, fmt.Errorf("connect to daemon: is daemon running?")
}
