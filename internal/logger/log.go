package logger

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type LogMsg string

type Logger struct {
	prevLogs []string
	ch       chan string
}

func NewLogger() *Logger {
	return &Logger{ch: make(chan string, 100)} // buffered channel
}

func (l *Logger) Add(msg string) {
	l.prevLogs = append(l.prevLogs, fmt.Sprintf("[%s] %s",
		time.Now().Format("15:04:05"), string(msg)))
	select {
	case l.ch <- strings.Join(l.prevLogs, "\n"):
	default:
	}
}

func (l *Logger) Listen() tea.Cmd {
	return func() tea.Msg {
		msg := <-l.ch
		return LogMsg(msg)
	}
}

func (l *Logger) Get() string {
	return strings.Join(l.prevLogs, "\n")
}
