package cmd

import (
	"fmt"
	"log"
	"time"
)

const (
	RFC3339ms = "2006-01-02T15:04:05.000Z07:00"
)

func SetLogRFC3339() {
	log.SetFlags(0)
	log.SetOutput(new(logWriter))
}

type logWriter struct{}

func (writer *logWriter) Write(bts []byte) (int, error) {
	// time.RFC3339
	return fmt.Print(time.Now().Format(RFC3339ms) + " " + string(bts))
}
