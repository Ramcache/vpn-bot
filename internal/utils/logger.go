package utils

import "log"

type Logger struct{}

func NewLogger() *Logger {
	return &Logger{}
}

func (l *Logger) Info(msg string) {
	log.Println("[INFO]", msg)
}

func (l *Logger) Error(msg string) {
	log.Println("[ERROR]", msg)
}
