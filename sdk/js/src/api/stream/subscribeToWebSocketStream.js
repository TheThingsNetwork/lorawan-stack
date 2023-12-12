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
 * @param {string} endpoint - The stream endpoint.
 * @param {number} timeout - The timeout for the stream.
 *
 * @example
 * (async () => {
 *    const stream = await stream(
 *      { identifiers: [{ application_ids: { application_id: 'my-app' }}]},
 *      'http://localhost:8080',
 *      '/api/v3',
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
export default async (
  payload,
  baseUrl,
  endpoint = '/console/internal/events/',
  timeout = 10000,
) => {
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
  const url = baseUrl + endpoint

  await Promise.race([
    new Promise(async (resolve, reject) => {
      // Add the new subscription to the subscriptions object.
      // Also add the resolver function to the subscription object to be able
      // to resolve the promise after the subscription confirmation message.
      subscriptions = {
        ...subscriptions,
        [subscriptionId]: { ...initialListeners, url, _resolver: resolve },
      }

      try {
        const token = new Token().get()
        const tokenParsed = typeof token === 'function' ? `${(await token()).access_token}` : token
        const baseUrlParsed = baseUrl.replace('http', 'ws')

        // Open up the WebSocket connection if it doesn't exist.
        if (!wsInstances[url]) {
          wsInstances[url] = new WebSocket(`${baseUrlParsed}${endpoint}`, [
            'ttn.lorawan.v3.console.internal.events.v1',
            `ttn.lorawan.v3.header.authorization.bearer.${tokenParsed}`,
          ])

          // Broadcast connection errors to all listeners.
          wsInstances[url].addEventListener('error', error => {
            Object.values(subscriptions)
              .filter(s => s.url === url)
              .forEach(s => notify(s[EVENTS.ERROR], new Error(error)))
            // The error is an error event, but we should only throw proper errors.
            // It has an optional error code that we could use to map to a proper error.
            // However, the error codes are optional and not always used.
            reject(new Error('Error in WebSocket connection'))
          })

          // Event listener for 'close'
          wsInstances[url].addEventListener('close', closeEvent => {
            // TODO: Handle close event codes.
            // https://github.com/TheThingsNetwork/lorawan-stack/issues/6752

            delete wsInstances[url]
            Object.values(subscriptions)
              .filter(s => s.url === url)
              .forEach(s => notify(s[EVENTS.CLOSE], closeRequested))

            if (closeRequested) {
              resolve()
            } else {
              reject(
                new Error(`WebSocket connection closed unexpectedly with code ${closeEvent.code}`),
              )
            }
          })

          // After the WebSocket connection is open, add the event listeners.
          // Wait for the subscription confirmation message before resolving.
          wsInstances[url].addEventListener('message', ({ data }) => {
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
              if (!Object.values(subscriptions).some(s => s.url === url)) {
                wsInstances[url].close()
              }
            }
          })
        }

        if (wsInstances[url] && wsInstances[url].readyState === WebSocket.OPEN) {
          // If the WebSocket connection is already open, only add the subscription.
          wsInstances[url].send(subscriptionPayload)
        } else if (wsInstances[url] && wsInstances[url].readyState === WebSocket.CONNECTING) {
          // Otherwise wait for the connection to open and then add the subscription.
          wsInstances[url].addEventListener('open', () => {
            wsInstances[url].send(subscriptionPayload)
          })
        }
      } catch (error) {
        const err = error instanceof Error ? error : new Error(error)
        Object.values(subscriptions)
          .filter(s => s.url === url)
          .forEach(s => notify(s[EVENTS.ERROR], err))
        reject(err)
      }
    }),
    new Promise((resolve, reject) => setTimeout(() => reject(new Error('timeout')), timeout)),
  ])

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
      if (wsInstances[url]) {
        closeRequested = true
        wsInstances[url].send(unsubscribePayload)

        // Wait for the server to confirm the unsubscribe.
        return new Promise(resolve => {
          wsInstances[url].addEventListener('message', ({ data }) => {
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
