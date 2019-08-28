// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package log

import (
	"errors"
	"strings"
)

// ErrInvalidLevel indicates an invalid log level
var ErrInvalidLevel = errors.New("Invalid log level")

// Level is the level of logging.
type Level int8

const (
	// invalid is an invalid log level.
	invalid Level = iota

	// DebugLevel is the log level for debug messages, usually turned of in production.
	DebugLevel

	// InfoLevel is the log level for informational messages.
	InfoLevel

	// WarnLevel is the log level for warnings.
	// Warnings are more important than info but do not need individual human review.
	WarnLevel

	// ErrorLevel is the log level for high priority error messages.
	// When everything is running smoothly, an application should not be logging error level messages.
	ErrorLevel

	// FatalLevel the log level for unrecoverable errors.
	// Using this level will exit the program.
	FatalLevel
)

// String implements fmt.Stringer.
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case FatalLevel:
		return "fatal"
	default:
		return "invalid"
	}
}

// ParseLevel parses a string into a log level or returns an error otherwise.
func ParseLevel(str string) (Level, error) {
	switch strings.ToLower(str) {
	case "debug":
		return DebugLevel, nil
	case "info":
		return InfoLevel, nil
	case "warn":
		return WarnLevel, nil
	case "error":
		return ErrorLevel, nil
	case "fatal":
		return FatalLevel, nil
	default:
		return invalid, ErrInvalidLevel
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

// UnmarshalConfigString implements config.Configurable.
func (l *Level) UnmarshalConfigString(str string) error {
	return l.UnmarshalText([]byte(str))
}
