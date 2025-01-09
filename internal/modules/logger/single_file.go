package logger

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
)

type SingleFileWriter struct {
	Filepath string
	Filename string
	file     *os.File
	log      *log.Logger
	mu       sync.Mutex
	debug    bool
}

func NewSingleFileWriter(filepath string, filename string, debug bool) Writer {
	fw := &SingleFileWriter{
		Filepath: filepath,
		Filename: filename,
	}

	if _, err := os.Stat(fw.Filepath); os.IsNotExist(err) {
		os.MkdirAll(fw.Filepath, os.ModePerm)
	}

	fw.file = fw.openFile()
	fw.log = log.New(fw.file, "", log.Ldate|log.Ltime)
	fw.log.SetOutput(fw.file)
	fw.debug = debug

	return fw
}

func (fw *SingleFileWriter) openFile() (File *os.File) {
	logfile, err := os.OpenFile(fw.constructFileName(), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)

	if err != nil {
		fmt.Printf("Unable to open file, err: %v", err)
		return
	}

	return logfile
}

func (fw *SingleFileWriter) constructFileName() string {
	return fmt.Sprintf("%v/%v.log", fw.Filepath, fw.Filename)
}

func (fw *SingleFileWriter) Write(message string, stackTrace []byte, payload map[string][]byte) {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	if fw.debug {
		callerLineCode := ""
		var stackTraceMessages []string

		if len(stackTrace) > 0 {
			stackTraceMessages = strings.Split(string(stackTrace), "\n")
			stackTraceMessages = stackTraceMessages[5:]
			callerLineCode = strings.Split(stackTraceMessages[1], "+")[0]
			message += fmt.Sprintf("\n at %v \n", strings.TrimSpace(callerLineCode))
		}

		if len(payload) > 0 {
			message += "\n[Payload]:\n"
			for key, val := range payload {
				message += fmt.Sprintf("Type: %v\n", key)
				message += fmt.Sprintf("Value: %v\n", string(val))
			}
		}

		if len(stackTrace) > 0 {
			message += "\n[StackTrace]:\n"
			message += strings.Join(stackTraceMessages, "\n")
		}
	} else {
		stackTraceMessages := strings.Split(string(stackTrace), "\n")
		message = fmt.Sprintf("%s %s", strings.TrimSpace(stackTraceMessages[10]), message)
	}

	fw.log.Println(message)
}

func (fw *SingleFileWriter) Close() {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	err := fw.file.Sync()
	if err != nil && !errors.Is(err, os.ErrClosed) {
		panic(err)
	}

	err = fw.file.Close()
	if err != nil && !errors.Is(err, os.ErrClosed) {
		panic(err)
	}
}
