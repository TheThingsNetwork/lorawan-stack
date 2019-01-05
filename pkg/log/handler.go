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

// Handler is the interface of things that can handle log entries.
type Handler interface {
	HandleLog(Entry) error
}

// NoopHandler is a handler that does nothing.
var NoopHandler = &noopHandler{}

// noopHandler is a handler that does nothing.
type noopHandler struct{}

// HandleLog implements Handler.
func (h *noopHandler) HandleLog(Entry) error {
	return nil
}

// HandlerFunc is a function that implements Handler.
type HandlerFunc func(Entry) error

// HandleLog implements Handler.
func (fn HandlerFunc) HandleLog(e Entry) error {
	return fn(e)
}
