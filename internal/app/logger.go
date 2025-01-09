package app

import (
	"fmt"

	"be20250107/internal/config"
	"be20250107/internal/modules/logger"
)

func NewLogger(cc config.LogConfig, debug bool, appName string) (*logger.Logger, error) {
	writers := make(map[logger.Level][]logger.Writer)

	for _, writerConfig := range cc.Writers {
		var writer logger.Writer
		switch writerConfig.Driver {
		case "rotating-file":
			writer = logger.NewRotatingFileWriter(writerConfig.LogRotatingFileWriterConfig.Filepath, fmt.Sprintf("%s.%s", appName, writerConfig.LogRotatingFileWriterConfig.Filename), debug)
		case "single-file":
			writer = logger.NewSingleFileWriter(writerConfig.LogSingleFileWriterConfig.Filepath, fmt.Sprintf("%s.%s", appName, writerConfig.LogSingleFileWriterConfig.Filename), debug)

		default:
			return nil, fmt.Errorf("unsupported log driver: %s", writerConfig.Driver)
		}

		if len(writerConfig.LogLevel) == 0 {
			for _, level := range logger.Levels {
				writers[level] = append(writers[level], writer)
			}
		} else {
			for _, level := range writerConfig.LogLevel {
				l := logger.Level(level)
				writers[l] = append(writers[l], writer)
			}
		}
	}

	return &logger.Logger{
		Writers: writers,
	}, nil
}
