// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package log

// Option for the logger
type Option func(*Logger) error

// WithHandler sets the handler on the logger.
func WithHandler(handler Handler) Option {
	return func(logger *Logger) error {
		logger.mutex.Lock()
		defer logger.mutex.Unlock()

		if handler != nil {
			logger.Handler = handler
		}
		return nil
	}
}

// WithLevel sets the level on the logger.
func WithLevel(level Level) Option {
	return func(logger *Logger) error {
		if level != invalid {
			logger.Level = level
		}

		return nil
	}
}
