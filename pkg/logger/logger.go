/*
 * Copyright (c) 2024 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package logger

import (
	"fmt"
	"io"
	"os/exec"

	"github.com/sirupsen/logrus"
)

var (
	log = New()
)

type LogFormatter struct{}

func (s *LogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	msg := fmt.Sprintf("%s cbuild: %s\n", entry.Level.String(), entry.Message)
	return []byte(msg), nil
}

// CustomLogger wraps logrus.Logger and extends the Error method
type CustomLogger struct {
	*logrus.Logger
}

func (cl *CustomLogger) Format(entry *logrus.Entry) ([]byte, error) {
	msg := fmt.Sprintf("%s cbuild: %s\n", entry.Level.String(), entry.Message)
	return []byte(msg), nil
}

// New creates a new instance of CustomLogger
func New() *CustomLogger {
	logger := &CustomLogger{logrus.StandardLogger()}
	logger.SetFormatter(new(LogFormatter))
	return logger
}

// Error method overrides logrus.Error with additional custom logic
func Error(args ...interface{}) {
	for _, arg := range args {
		switch arg.(type) {
		case *exec.ExitError:
			logrus.Info(arg)
		default:
			logrus.Error(arg)
		}
	}
}

// Wrapping the standard functions
func Info(args ...interface{}) {
	log.Info(args...)
}

func Warn(args ...interface{}) {
	log.Warn(args...)
}

func Debug(args ...interface{}) {
	log.Debug(args...)
}

func SetLevel(level logrus.Level) {
	log.SetLevel(level)
}

func GetLevel() logrus.Level {
	return log.GetLevel()
}

func SetFormatter(formatter logrus.Formatter) {
	log.SetFormatter(formatter)
}

func SetOutput(out io.Writer) {
	log.SetOutput(out)
}

func StandardLogger() *logrus.Logger {
	return logrus.StandardLogger()
}
