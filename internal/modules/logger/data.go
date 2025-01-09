package logger

import (
	"encoding/json"
)

type LogData interface {
	// Ident returns the identifier for this data.
	// This value is used to differentiate between data sent to the log function.
	Ident() string
	Serialize() []byte
}
type StackTraceData struct {
	StackTrace []byte
}

func (d StackTraceData) Ident() string {
	return "stack-trace-data"
}

func (d StackTraceData) Serialize() []byte {
	return []byte{}
}

type PayloadData struct {
	Data any
}

func (pd PayloadData) Ident() string {
	return "payload-data"
}

func (pd PayloadData) Serialize() []byte {
	payload, err := json.Marshal(pd)
	if err != nil {
		panic(err)
	}
	return payload
}

type ContextData struct {
	UserID string
}

func (cd ContextData) Ident() string {
	return "context-data"
}

func (cd ContextData) Serialize() []byte {
	return []byte{}
}
