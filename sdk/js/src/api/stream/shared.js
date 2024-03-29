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

export const notify = (listener, ...args) => {
  if (typeof listener === 'function') {
    listener(...args)
  }
}

export const EVENTS = Object.freeze({
  MESSAGE: 'message',
  ERROR: 'error',
  CLOSE: 'close',
})

export const MESSAGE_TYPES = Object.freeze({
  SUBSCRIBE: 'subscribe',
  UNSUBSCRIBE: 'unsubscribe',
  PUBLISH: 'publish',
  ERROR: 'error',
})

export const INITIAL_LISTENERS = Object.freeze(
  Object.values(EVENTS).reduce((acc, curr) => ({ ...acc, [curr]: {} }), {}),
)

export const newQueuedListeners = listeners => {
  const queue = []
  let open = false
  const queuedListeners = Object.values(EVENTS).reduce(
    (acc, curr) => ({
      ...acc,
      [curr]: (...args) => {
        if (open) {
          notify(listeners[curr], ...args)
        } else {
          queue.push([curr, args])
        }
      },
    }),
    {},
  )
  return [
    () => {
      open = true
      for (const [event, args] of queue) {
        notify(listeners[event], ...args)
      }
      queue.splice(0, queue.length)
    },
    queuedListeners,
  ]
}
