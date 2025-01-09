package logger

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

var now = time.Now

type RotatingFileWriter struct {
	Filepath     string
	Filename     string
	currFilename string
	file         *os.File
	log          *log.Logger
	mu           sync.Mutex
	debug        bool
}

func NewRotatingFileWriter(filepath string, filename string, debug bool) Writer {
	fw := &RotatingFileWriter{
		Filepath: filepath,
		Filename: filename,
	}

	if _, err := os.Stat(fw.Filepath); os.IsNotExist(err) {
		os.MkdirAll(fw.Filepath, os.ModePerm)
	}

	file := fw.openFile()
	log := log.New(file, "", log.Ldate|log.Ltime)

	fw.file = file
	fw.log = log
	fw.log.SetOutput(fw.file)
	fw.debug = debug

	return fw
}

func (fw *RotatingFileWriter) openFile() (File *os.File) {
	fw.currFilename = fw.constructFileName()
	logfile, err := os.OpenFile(fw.currFilename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)

	if err != nil {
		fmt.Printf("Unable to open file, err: %v", err)
		return
	}

	return logfile
}

func (fw *RotatingFileWriter) rotateFile() {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	if fw.file != nil {
		fw.file.Close()
	}

	fw.file = fw.openFile()
	fw.log.SetOutput(fw.file)
}

func (fw *RotatingFileWriter) constructFileName() string {
	return fmt.Sprintf("%v/%v-%v.log", fw.Filepath, now().Format("2006-01-02"), fw.Filename)
}

func (fw *RotatingFileWriter) Write(message string, stackTrace []byte, payload map[string][]byte) {
	if fw.constructFileName() != fw.currFilename {
		fw.rotateFile()
	}

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

func (fw *RotatingFileWriter) Close() {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	err := fw.file.Sync()
	if err != nil && !errors.Is(err, os.ErrClosed) {
		panic(err)
	}

	err = fw.file.Close()
	//fmt.Println(err)
	if err != nil && !errors.Is(err, os.ErrClosed) {
		panic(err)
	}
}
