package logger

import (
	"fmt"
	"log"
	"runtime/debug"
)

type Level string

const (
	Verbose Level = "Verbose"
	Info    Level = "Info"
	Warning Level = "Warning"
	Error   Level = "Error"
	Fatal   Level = "Fatal"
)

var Levels = []Level{Verbose, Info, Warning, Error, Fatal}

type Logger struct {
	Writers map[Level][]Writer
}

type Writer interface {
	Write(message string, stackTrace []byte, payload map[string][]byte)
	Close()
}

func (l Logger) Verbose(message string, data ...LogData) {
	message = fmt.Sprint("[Verbose] ", message)
	stackTrace := []byte{}
	payload := make(map[string][]byte)

	for _, d := range data {
		if st, ok := d.(StackTraceData); ok {
			stackTrace = st.StackTrace
		} else {
			payload[fmt.Sprintf("%T", d)] = d.Serialize()
		}
	}
	for _, w := range l.Writers[Error] {
		w.Write(message, stackTrace, payload)
	}
}

func (l Logger) Info(message string, data ...LogData) {
	message = fmt.Sprint("[Info] ", message)
	stackTrace := []byte{}
	payload := make(map[string][]byte)

	for _, d := range data {
		if st, ok := d.(StackTraceData); ok {
			stackTrace = st.StackTrace
		} else {
			payload[fmt.Sprintf("%T", d)] = d.Serialize()
		}
	}
	for _, w := range l.Writers[Error] {
		w.Write(message, stackTrace, payload)
	}
}

func (l Logger) Warning(message string, data ...LogData) {
	message = fmt.Sprint("[Warning] ", message)
	stackTrace := []byte{}
	payload := make(map[string][]byte)

	for _, d := range data {
		if st, ok := d.(StackTraceData); ok {
			stackTrace = st.StackTrace
		} else {
			payload[fmt.Sprintf("%T", d)] = d.Serialize()
		}
	}
	for _, w := range l.Writers[Error] {
		w.Write(message, stackTrace, payload)
	}
}

func (l Logger) Errorf(format string, a ...any) {
	l.error(fmt.Sprintf(format, a...))
}

func (l Logger) Error(message string, a ...any) {
	l.error(message + fmt.Sprint(a...))
}

func (l Logger) error(message string, data ...LogData) {
	message = fmt.Sprint("[Error] ", message)
	log.Println(message) // output to console too

	stackTrace := debug.Stack()
	payload := make(map[string][]byte)

	for _, d := range data {
		if st, ok := d.(StackTraceData); ok {
			stackTrace = st.StackTrace
		} else {
			payload[fmt.Sprintf("%T", d)] = d.Serialize()
		}
	}
	for _, w := range l.Writers[Error] {
		w.Write(message, stackTrace, payload)
	}
}

func (l Logger) Fatal(message string, data ...LogData) {
	message = fmt.Sprint("[Fatal] ", message)
	stackTrace := debug.Stack()
	payload := make(map[string][]byte)

	for _, d := range data {
		if st, ok := d.(StackTraceData); ok {
			stackTrace = st.StackTrace
		} else {
			payload[fmt.Sprintf("%T", d)] = d.Serialize()
		}
	}
	for _, w := range l.Writers[Error] {
		w.Write(message, stackTrace, payload)
	}
}
