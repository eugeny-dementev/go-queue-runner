package queuerunner

import (
	"log"
)

type Logger interface {
	Info(message string)
	SetContext(context string)
	Error(err error)
}

type stdLogger struct{}

func (logger stdLogger) Info(message string) {
	log.Print(message)
}

func (logger stdLogger) SetContext(_ string) {}

func (logger stdLogger) Error(err error) {
	log.Print(err)
}

func defaultLogger() Logger {
	return stdLogger{}
}
