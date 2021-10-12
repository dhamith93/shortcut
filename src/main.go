package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dhamith93/shortcut/internal/server"
)

func main() {
	f, err := os.OpenFile("shortcut.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	handler := server.Handler{}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-ch
		server.Shutdown(handler, ctx)
		signal.Reset(syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		cancel()
		os.Exit(0)
	}()
	server.Run(handler)
}
