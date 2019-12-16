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

class EventHandler {
  constructor() {
    this.EVENTS = Object.freeze({
      WARNING: 'warning',
      // Add more here as we go
    })

    this.eventHandlers = {}
    this.dispatchEvent = function(event, payload) {
      if (this.eventHandlers[event]) {
        for (const handler of this.eventHandlers[event]) {
          handler(payload)
        }
      }
    }.bind(this)

    this.subscribe = function(event, handler) {
      if (!Object.values(this.EVENTS).includes(event)) {
        throw new Error(`Cannot subscribe to unsupported event type "${event}"`)
      }
      this.eventHandlers = {
        ...this.eventHandlers,
        [event]: this.eventHandlers[event] ? [...this.eventHandlers[event], handler] : [handler],
      }
    }.bind(this)

    this.unsubscribe = function(event) {
      if (!Object.values(this.EVENTS).includes(event)) {
        throw new Error(`Cannot unsubscribe from unsupported event type "${event}"`)
      }
      if (this.eventHandlers[event]) {
        delete this.eventHandlers[event]
      }
    }.bind(this)
  }
}

export default new EventHandler()
