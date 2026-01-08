package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/aster-void/webhooker/internal/file"
	"github.com/aster-void/webhooker/internal/ipc"
	"github.com/aster-void/webhooker/internal/receiver"
	"github.com/aster-void/webhooker/internal/router"
)

const (
	readTimeout    = 10 * time.Second
	writeTimeout   = 5 * time.Second
	maxHeaderSize  = 8 << 10 // 8KB
	defaultDataDir = "/var/lib/webhooker"
)

func main() {
	if len(os.Args) >= 2 && os.Args[1] == "daemon" {
		runDaemon()
		return
	}
	if err := ipc.RunClient(getSocketPath()); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func runDaemon() {
	port := os.Getenv("WEBHOOKER_PORT")
	if port == "" {
		port = "8080"
	}
	logDir := getLogDir()

	// Channels
	recvCh := make(chan receiver.Message, 100)
	fileCh := make(chan []byte, 100)

	// File writer
	fileWriter, err := file.NewFile(logDir, fileCh)
	if err != nil {
		log.Fatalf("failed to create file writer: %v", err)
	}
	defer fileWriter.Close()
	go fileWriter.Run()

	// Router
	r := router.New(recvCh)
	r.LoadRoutes(fileCh)
	go r.Run()

	// IPC server
	ipcServer, err := ipc.NewServer(getSocketPath(), getDomain(), r.Register, r.Unregister)
	if err != nil {
		log.Fatalf("failed to create IPC server: %v", err)
	}
	defer ipcServer.Close()
	go ipcServer.Run()

	// Receiver
	recv := receiver.New(recvCh)

	srv := &http.Server{
		Addr:           ":" + port,
		Handler:        recv,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: maxHeaderSize,
	}

	done := make(chan struct{})
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
		<-sigCh

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("shutdown error: %v", err)
		}
		close(done)
	}()

	log.Printf("starting server on :%s", port)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
	<-done
}

func getDataDir() string {
	if dir := os.Getenv("WEBHOOKER_DATA_DIR"); dir != "" {
		return dir
	}
	if os.Getuid() == 0 {
		return defaultDataDir
	}
	if xdg := os.Getenv("XDG_STATE_HOME"); xdg != "" {
		return xdg + "/webhooker"
	}
	if home := os.Getenv("HOME"); home != "" {
		return home + "/.local/state/webhooker"
	}
	return defaultDataDir
}

func getLogDir() string {
	if dir := os.Getenv("WEBHOOKER_LOG_DIR"); dir != "" {
		return dir
	}
	return getDataDir()
}

func getSocketPath() string {
	if path := os.Getenv("WEBHOOKER_SOCKET"); path != "" {
		return path
	}
	return "/run/webhooker/webhooker.sock"
}

func getDomain() string {
	return strings.TrimSuffix(os.Getenv("WEBHOOKER_DOMAIN"), "/")
}
