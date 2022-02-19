package adapter

import (
	"fmt"

	"github.com/ethereum/go-ethereum/log"
)

type Logger struct {
}

func (l *Logger) Info(a ...interface{}) {
	log.Info("", a...)
}

func (l *Logger) Infof(format string, a ...interface{}) {
	log.Info(fmt.Sprintf(format, a...))
}

func (l *Logger) Debug(a ...interface{}) {
	log.Debug("", a...)
}

func (l *Logger) Debugf(format string, a ...interface{}) {
	log.Debug(fmt.Sprintf(format, a...))
}

func (l *Logger) Fatal(a ...interface{}) {
	log.Crit("", a...)
}

func (l *Logger) Fatalf(format string, a ...interface{}) {
	log.Crit(fmt.Sprintf(format, a...))
}

func (l *Logger) Error(a ...interface{}) {
	log.Warn("", a...)
}

func (l *Logger) Errorf(format string, a ...interface{}) {
	l.Error(fmt.Sprintf(format, a...))
}
