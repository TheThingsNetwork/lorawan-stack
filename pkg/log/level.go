// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package log

import (
	"errors"
	"strings"
)

var ErrInvalidLevel = errors.New("Invalid log level")

// Level is the level of logging.
type Level int8

const (
	// Debug is the log level for debug messages, usually turned of in production.
	Debug Level = iota

	// Info is the log level for informational messages.
	Info

	// Warn is the log level for warnings.
	// Warnings are more important than info but do not need individual human review.
	Warn

	// Error is the log level for high priority error messages.
	// When everything is running smoothly, an application should not be logging error level messages.
	Error

	// Fatal the log level for unrecoverable errors.
	Fatal Level = iota
)

// String implements fmt.Stringer.
func (l Level) String() string {
	switch l {
	case Debug:
		return "debug"
	case Info:
		return "info"
	case Warn:
		return "warn"
	case Error:
		return "error"
	case Fatal:
		return "fatal"
	default:
		return "invalid"
	}
}

// ParseLevel parses a string into a log level.
func ParseLevel(str string) (Level, error) {
	switch strings.ToLower(str) {
	case "debug":
		return Debug, nil
	case "info":
		return Info, nil
	case "warn":
		return Warn, nil
	case "error":
		return Error, nil
	case "fatal":
		return Fatal, nil
	default:
		return Fatal, ErrInvalidLevel
	}
}

// MarshalText implments encoding.TextMarshaller.
func (l Level) MarshalText() ([]byte, error) {
	return []byte(l.String()), nil
}

// UnmarshalText implments encoding.TextMarshaller.
func (l *Level) UnmarshalText(text []byte) error {
	level, err := ParseLevel(string(text))
	if err != nil {
		return err
	}

	*l = level

	return nil
}
