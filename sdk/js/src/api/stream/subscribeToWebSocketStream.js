// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

import traverse from 'traverse'

import Token from '../../util/token'

import { notify, newQueuedListeners, EVENTS, MESSAGE_TYPES, INITIAL_LISTENERS } from './shared'

export class ConnectionError extends Error {
  constructor(message) {
    super(message)
    this.name = 'ConnectionError'
  }
}

export class ConnectionClosedError extends ConnectionError {
  constructor(message, code) {
    super(message)
    this.name = 'ConnectionClosedError'
    this.code = code
  }
}

export class ConnectionTimeoutError extends ConnectionError {
  constructor(message) {
    super(message)
    this.name = 'ConnectionTimeoutError'
  }
}

export class ProtocolError extends Error {
  constructor(error) {
    super(error.message)
    this.name = 'ProtocolError'
    this.error = error
  }
}

const newSubscription = (
  unsubscribe,
  originalListeners,
  resolveSubscribe,
  rejectSubscribe,
  resolveClose,
) => {
  let closeRequested = false
  const [open, listeners] = newQueuedListeners(originalListeners)
  const externalSubscription = {
    open,
    close: () => {
      closeRequested = true
      return unsubscribe()
    },
  }
  return {
    onError: err => {
      notify(listeners[EVENTS.ERROR], err)

      rejectSubscribe(err)
    },
    onClose: closeEvent => {
      notify(listeners[EVENTS.CLOSE], closeRequested)

      rejectSubscribe(new ConnectionClosedError('WebSocket connection closed', closeEvent.code))
      resolveClose()
    },
    onMessage: dataParsed => {
      if (dataParsed.type === MESSAGE_TYPES.SUBSCRIBE) {
        resolveSubscribe(externalSubscription)
      }

      if (dataParsed.type === MESSAGE_TYPES.ERROR) {
        const err = new ProtocolError(dataParsed.error)
        notify(listeners[EVENTS.ERROR], err)

        rejectSubscribe(err)
      }

      if (dataParsed.type === MESSAGE_TYPES.PUBLISH) {
        notify(listeners[EVENTS.MESSAGE], dataParsed.event)
      }

      if (dataParsed.type === MESSAGE_TYPES.UNSUBSCRIBE) {
        notify(listeners[EVENTS.CLOSE], closeRequested)

        resolveClose()
      }
    },
  }
}

const newInstance = (wsInstance, onClose) => {
  const subscriptions = {}

  // Broadcast connection errors to all subscriptions.
  wsInstance.addEventListener('error', () => {
    const err = new ConnectionError('WebSocket connection error')
    for (const subscription of Object.values(subscriptions)) {
      subscription.onError(err)
    }
  })

  // Broadcast connection closure to all subscriptions.
  wsInstance.addEventListener('close', closeEvent => {
    // TODO: Handle close event codes.
    // https://github.com/TheThingsNetwork/lorawan-stack/issues/6752
    for (const subscription of Object.values(subscriptions)) {
      subscription.onClose(closeEvent)
    }
    onClose()
  })

  // Broadcast messages to the correct subscription.
  wsInstance.addEventListener('message', ({ data }) => {
    const dataParsed = JSON.parse(data)
    const sid = dataParsed.id
    const subscription = traverse(subscriptions).get([sid]) || null

    if (!subscription) {
      return
    }

    subscription.onMessage(dataParsed)

    if (dataParsed.type === MESSAGE_TYPES.UNSUBSCRIBE) {
      delete subscriptions[sid]
    }
  })

  return {
    subscribe: (sid, subscribePayload, unsubscribePayload, listeners, resolve, reject) => {
      if (sid in subscriptions) {
        throw new Error(`Subscription with ID ${sid} already exists`)
      }

      // The `unsubscribed` promise is used in order to guarantee that calls to the `close` method
      // of the subscription finish _after_ the closure events have been emitted. Callers can expect
      // that after `close` resolves, no further events will be emitted.
      let resolveClose = null
      const unsubscribed = new Promise(resolve => {
        resolveClose = resolve
      })
      let unsubscribeCalled = false
      const unsubscribe = () => {
        if (!unsubscribeCalled) {
          unsubscribeCalled = true

          if (wsInstance.state === WebSocket.open) {
            wsInstance.send(unsubscribePayload)
          }
        }
        return unsubscribed
      }

      const subscription = newSubscription(unsubscribe, listeners, resolve, reject, resolveClose)
      subscriptions[sid] = subscription

      if (wsInstance.readyState === WebSocket.OPEN) {
        // If the WebSocket connection is already open, only add the subscription.
        wsInstance.send(subscribePayload)
      } else if (wsInstance.readyState === WebSocket.CONNECTING) {
        // Otherwise wait for the connection to open and then add the subscription.
        const onOpen = () => {
          wsInstance.send(subscribePayload)
          wsInstance.removeEventListener('open', onOpen)
        }
        wsInstance.addEventListener('open', onOpen)
      } else {
        delete subscriptions[sid]
        throw new Error('WebSocket connection is closed')
      }
    },
  }
}

const newStore = () => {
  const connections = {}
  return {
    getInstance: url => traverse(connections).get([url]),
    setInstance: (url, wsInstance) =>
      traverse(connections).set(
        [url],
        newInstance(wsInstance, () => delete connections[url]),
      ),
  }
}

const state = newStore()

/**
 * Opens a new stream.
 *
 * Implementation guarantees:
 * - No events will be sent to the listeners before the `open` function is called.
 * - No events will be sent to the listeners after the `close` function is called.
 * - The `close` function will resolve after all events have been sent to the listeners.
 * - The `close` function does not throw any errors.
 * - The `close` function can be called multiple times.
 * - The `open` function can be called multiple times.
 * - The `open` function does not throw any errors, as long as the event listeners do not throw.
 * - No `message` event will follow an `error` event.
 * - No `message` event will follow a `close` event.
 * - No `error` event will follow a `close` event.
 * - No `error` event will follow another `error` event.
 * - No `close` event will follow another `close` event.
 *
 * @async
 * @param {object} payload -  - The body of the initial request.
 * @param {string} baseUrl - The stream baseUrl.
 * @param {object} listeners - The listeners object.
 * @param {string} endpoint - The stream endpoint.
 * @param {number} timeout - The connection timeout for the stream.
 *
 * @example
 * (async () => {
 *    const stream = await stream(
 *      { identifiers: [{ application_ids: { application_id: 'my-app' }}]},
 *      'http://localhost:8080',
 *      {
 *        message: ({ data }) => console.log('received data', JSON.parse(data)),
 *        error: error => console.log(error),
 *        close: wasClientRequest => console.log(wasClientRequest ? 'conn closed by client' : 'conn closed by server'),
 *      },
 *    )
 *
 *    // Start the stream in order to start dispatching events.
 *    stream.open()
 *
 *    // Close the stream after 20 s.
 *    setTimeout(() => stream.close(), 20000)
 * })()
 *
 * @returns {object} The stream subscription object the `open` function to start sending events to the listeners and
 * the `close` function to close the stream.
 */
export default async (
  payload,
  baseUrl,
  listeners,
  endpoint = '/console/internal/events/',
  timeout = 10000,
) => {
  for (const eventName of Object.keys(listeners)) {
    if (!Object.values(EVENTS).includes(eventName)) {
      throw new Error(
        `${eventName} event is not supported. Should be one of: message, error or close`,
      )
    }
  }
  const filledListeners = { ...INITIAL_LISTENERS, ...listeners }

  const subscriptionId = Math.floor(Math.random() * Number.MAX_SAFE_INTEGER)
  const subscriptionPayload = JSON.stringify({
    type: MESSAGE_TYPES.SUBSCRIBE,
    id: subscriptionId,
    ...payload,
  })
  const unsubscribePayload = JSON.stringify({
    type: MESSAGE_TYPES.UNSUBSCRIBE,
    id: subscriptionId,
  })
  const url = baseUrl + endpoint
  const token = new Token().get()
  const tokenParsed = typeof token === 'function' ? `${(await token()).access_token}` : token
  const baseUrlParsed = baseUrl.replace('http', 'ws')

  const subscribe = new Promise((resolve, reject) => {
    let instance = state.getInstance(url)
    // Open up the WebSocket connection if it doesn't exist.
    if (!instance) {
      instance = state.setInstance(
        url,
        new WebSocket(`${baseUrlParsed}${endpoint}`, [
          'ttn.lorawan.v3.console.internal.events.v1',
          `ttn.lorawan.v3.header.authorization.bearer.${tokenParsed}`,
        ]),
      )
    }

    // Add the new subscription to the subscriptions object.
    // Also add the resolver functions to the subscription object to be able
    // to resolve the promise after the subscription confirmation message.
    instance.subscribe(
      subscriptionId,
      subscriptionPayload,
      unsubscribePayload,
      filledListeners,
      resolve,
      reject,
    )
  })

  try {
    return await Promise.race([
      subscribe,
      new Promise((_resolve, reject) =>
        setTimeout(() => reject(new ConnectionTimeoutError('Timed out')), timeout),
      ),
    ])
  } catch (err) {
    subscribe.then(
      subscription => subscription.close(),
      () => {},
    )
    throw err
  }
}
