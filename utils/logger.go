package utils

import (
	"encoding/json"
	"log"
)

// Fields represents structured logging fields.
type Fields map[string]interface{}

// Logger is a minimal structured logging interface.
type Logger interface {
	Info(msg string, fields Fields)
	Error(msg string, fields Fields)
	With(fields Fields) Logger
}

// stdLogger is a basic implementation that prints JSON lines via the stdlib log package.
type stdLogger struct {
	base Fields
}

// NewStdLogger creates a new standard logger.
func NewStdLogger() Logger {
	return &stdLogger{base: Fields{}}
}

func (l *stdLogger) With(fields Fields) Logger {
	merged := Fields{}
	for k, v := range l.base {
		merged[k] = v
	}
	for k, v := range fields {
		merged[k] = v
	}
	return &stdLogger{base: merged}
}

func (l *stdLogger) Info(msg string, fields Fields) {
	l.log("info", msg, fields)
}

func (l *stdLogger) Error(msg string, fields Fields) {
	l.log("error", msg, fields)
}

func (l *stdLogger) log(level, msg string, fields Fields) {
	payload := Fields{"level": level, "msg": msg}
	// merge base fields
	for k, v := range l.base {
		payload[k] = v
	}
	// merge call fields
	for k, v := range fields {
		payload[k] = v
	}
	b, err := json.Marshal(payload)
	if err != nil {
		// fallback to simple print
		log.Printf("level=%s msg=\"%s\"", level, msg)
		return
	}
	log.Printf(string(b))
}
