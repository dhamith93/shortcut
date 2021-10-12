package logger

import (
	"log"
	"os"
	"strings"
)

// Log logs to given log file
func Log(prefix string, msg string) {
	log.Println(strings.ToUpper(prefix) + " " + msg)
}

func Fatal(msg string) {
	Log("error", msg)
	os.Exit(1)
}
