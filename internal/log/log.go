package log

import (
	"fmt"
	"log"
	"time"
)

// Printf logs a message with an ISO8601 timestamp prefix.
func Printf(format string, args ...any) {
	ts := time.Now().UTC().Format(time.RFC3339Nano)
	log.Printf("[%s] "+format, append([]any{ts}, args...)...)
}

// Println logs messages with an ISO8601 timestamp prefix.
func Println(args ...any) {
	ts := time.Now().UTC().Format(time.RFC3339Nano)
	log.Println(append([]any{fmt.Sprintf("[%s]", ts)}, args...)...)
}
