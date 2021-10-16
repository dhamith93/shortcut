package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	f, err := os.OpenFile("shortcut.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	handler := handler{}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-ch
		shutdown(handler, ctx)
		signal.Reset(syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		cancel()
		os.Exit(0)
	}()
	start(handler)
}

func start(handler handler) {
	cleanUp()
	handler.config = loadConfig("config.json")
	handler.fileList = getFileList()
	handler.clipboardItems = []clipboardItem{}
	handler.hub = newHub()
	go handler.hub.run()
	handler.handleRequests()
}

func shutdown(handler handler, ctx context.Context) {
	cleanUp()
	err := handler.shutdown(ctx)
	if err != nil {
		Log("error", err.Error())
	}
	Log("info", "Shortcut stopped")
}
