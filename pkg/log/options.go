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
