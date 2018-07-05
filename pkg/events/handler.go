// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package events

// Handler interface for event listeners.
type Handler interface {
	Notify(Event)
}

type handlerFunc struct {
	handler func(Event)
}

func (f handlerFunc) Notify(evt Event) { f.handler(evt) }

// HandlerFunc makes the func implement the Handler interface.
func HandlerFunc(handler func(Event)) Handler {
	return &handlerFunc{handler}
}

// Channel is a channel of Events that can be used as an event handler.
// The channel should be buffered, events will be dropped if the channel blocks.
// It is typically not safe to close this channel until you're absolutely sure
// that it is no longer registered as an event handler.
type Channel chan Event

// Notify implements the Handler interface.
func (ch Channel) Notify(evt Event) {
	select {
	case ch <- evt:
	default:
	}
}
