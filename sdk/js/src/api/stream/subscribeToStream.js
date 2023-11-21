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

import Token from '../../util/token'
import { warn } from '../../../../../pkg/webui/lib/log'

import { notify, EVENTS, MESSAGE_TYPES } from './shared'

const wsInstances = {}
let subscriptions = {}
const initialListeners = Object.values(EVENTS).reduce((acc, curr) => ({ ...acc, [curr]: {} }), {})

/**
 * Opens a new stream.
 *
 * @async
 * @param {object} payload -  - The body of the initial request.
 * @param {string} baseUrl - The stream baseUrl.
 *
 * @example
 * (async () => {
 *    const stream = await stream(
 *      { identifiers: [{ application_ids: { application_id: 'my-app' }}]},
 *      'http://localhost:8080/api/v3',
 *    )
 *
 *    // Add listeners to the stream.
 *    stream
 *      .on('open', () => console.log('conn opened'))
 *      .on('message', ({ data }) => console.log('received data', JSON.parse(data)))
 *      .on('error', error => console.log(error))
 *      .on('close', wasClientRequest => console.log(wasClientRequest ? 'conn closed by client' : 'conn closed by server'))
 *
 *     // Close the stream after 20 s.
 *    setTimeout(() => stream.close(), 20000)
 * })()
 *
 * @returns {object} The stream subscription object with the `on` function for
 * attaching listeners and the `close` function to close the stream.
 */
export default async (payload, baseUrl) => {
  const subscriptionId = Date.now()
  const subscriptionPayload = JSON.stringify({
    type: MESSAGE_TYPES.SUBSCRIBE,
    id: subscriptionId,
    ...payload,
  })
  const unsubscribePayload = JSON.stringify({
    type: MESSAGE_TYPES.UNSUBSCRIBE,
    id: subscriptionId,
  })
  let closeRequested = false

  await new Promise(async resolve => {
    // Add the new subscription to the subscriptions object.
    // Also add the resolver function to the subscription object to be able
    // to resolve the promise after the subscription confirmation message.
    subscriptions = {
      ...subscriptions,
      [subscriptionId]: { ...initialListeners, url: baseUrl, _resolver: resolve },
    }

    const token = new Token().get()
    const tokenParsed = typeof token === 'function' ? (await token()).access_token : token
    const baseUrlParsed = baseUrl.replace('http', 'ws')

    // Open up the WebSocket connection if it doesn't exist.
    if (!wsInstances[baseUrl]) {
      wsInstances[baseUrl] = new WebSocket(`${baseUrlParsed}/console/internal/events/`, [
        'ttn.lorawan.v3.console.internal.events.v1',
        `ttn.lorawan.v3.header.authorization.bearer.${tokenParsed}`,
      ])

      // Event listener for 'open'
      wsInstances[baseUrl].addEventListener('open', () => {
        wsInstances[baseUrl].send(subscriptionPayload)
      })

      // Broadcast connection errors to all listeners.
      wsInstances[baseUrl].addEventListener('error', error => {
        Object.values(subscriptions)
          .filter(s => s.url === baseUrl)
          .forEach(s => notify(s[EVENTS.ERROR], error))
        resolve()
      })

      // Event listener for 'close'
      wsInstances[baseUrl].addEventListener('close', () => {
        delete wsInstances[baseUrl]
      })

      // After the WebSocket connection is open, add the event listeners.
      // Wait for the subscription confirmation message before resolving.
      wsInstances[baseUrl].addEventListener('message', ({ data }) => {
        const dataParsed = JSON.parse(data)
        const listeners = subscriptions[dataParsed.id]

        if (!listeners) {
          warn('Message received for closed or unknown subscription with ID', dataParsed.id)

          return
        }

        if (dataParsed.type === MESSAGE_TYPES.SUBSCRIBE) {
          notify(listeners[EVENTS.OPEN])
          // Resolve the promise after the subscription confirmation message.
          listeners._resolver()
        }

        if (dataParsed.type === MESSAGE_TYPES.ERROR) {
          notify(listeners[EVENTS.ERROR], dataParsed)
        }

        if (dataParsed.type === MESSAGE_TYPES.PUBLISH) {
          notify(listeners[EVENTS.MESSAGE], dataParsed.event)
        }

        if (dataParsed.type === MESSAGE_TYPES.UNSUBSCRIBE) {
          notify(listeners[EVENTS.CLOSE], closeRequested)
          // Remove the subscription.
          delete subscriptions[dataParsed.id]
          if (!Object.values(subscriptions).some(s => s.url === baseUrl)) {
            wsInstances[baseUrl].close()
          }
        }
      })
    } else if (wsInstances[baseUrl] && wsInstances[baseUrl].readyState === WebSocket.OPEN) {
      // If the WebSocket connection is already open, only add the subscription.
      wsInstances[baseUrl].send(subscriptionPayload)
    }
  })

  // Return an observer object with the `on` and `close` functions for
  // the current subscription.
  return {
    on(eventName, callback) {
      if (!Object.values(EVENTS).includes(eventName)) {
        throw new Error(
          `${eventName} event is not supported. Should be one of: open, message, error or close`,
        )
      }
      subscriptions[subscriptionId][eventName] = callback

      return this
    },
    close: () => {
      if (wsInstances[baseUrl]) {
        closeRequested = true
        wsInstances[baseUrl].send(unsubscribePayload)

        // Wait for the server to confirm the unsubscribe.
        return new Promise(resolve => {
          wsInstances[baseUrl].addEventListener('message', ({ data }) => {
            const { type, id } = JSON.parse(data)
            if (id === subscriptionId && type === MESSAGE_TYPES.UNSUBSCRIBE) {
              resolve()
            }
          })
        })
      }
    },
  }
}
