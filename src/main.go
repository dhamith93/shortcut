package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/dhamith93/shortcut/internal/server"
)

func main() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-ch
		server.Shutdown(ctx)
		signal.Reset(syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		cancel()
	}()
	server.Run()
}
