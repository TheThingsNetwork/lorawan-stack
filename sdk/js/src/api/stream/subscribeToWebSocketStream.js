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
import { warn } from '../../../../../pkg/webui/lib/log'

import { notify, EVENTS, MESSAGE_TYPES } from './shared'

const initialListeners = Object.values(EVENTS).reduce((acc, curr) => ({ ...acc, [curr]: {} }), {})

const store = () => {
  const connections = {}

  return {
    getInstance: url => traverse(connections).get([url, 'instance']),
    setInstance: (url, instance) => {
      traverse(connections).set([url, 'instance'], instance)
      return instance
    },
    deleteInstance: url => {
      if (url in connections) {
        delete connections[url]
      }
    },
    getSubscriptions: url => Object.values(traverse(connections).get([url, 'subscriptions'] || {})),
    getSubscription: (url, sid) => traverse(connections).get([url, 'subscriptions', sid]) || null,
    setSubscription: (url, sid, subscription) => {
      const subs = traverse(connections).get([url, 'subscriptions']) || {}
      subs[sid] = subscription
      traverse(connections).set([url, 'subscriptions'], subs)
      return subs[sid]
    },
    markSubscriptionClosing: (url, sid) => {
      if (traverse(connections).has([url, 'subscriptions', sid])) {
        traverse(connections).set([url, 'subscriptions', sid, 'closeRequested'], true)
      }
    },
    getSubscriptionCount: url => {
      const subs = traverse(connections).get([url, 'subscriptions'])
      return subs ? Object.keys(subs).length : 0
    },
    deleteSubscription: (url, sid) => {
      const subscriptions = traverse(connections).get([url, 'subscriptions'])
      if (subscriptions && subscriptions[sid]) {
        delete subscriptions[sid]
      }
    },
  }
}

const state = store()

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
  let wsInstance = state.getInstance(url)

  await Promise.race([
    new Promise(async (resolve, reject) => {
      // Add the new subscription to the subscriptions object.
      // Also add the resolver functions to the subscription object to be able
      // to resolve the promise after the subscription confirmation message.
      if (state.getSubscription(url, subscriptionId) !== null) {
        reject(new Error('Subscription with the same ID already exists'))
      }
      state.setSubscription(url, subscriptionId, {
        ...initialListeners,
        url,
        resolve,
        reject,
        closeRequested: false,
      })

      try {
        const token = new Token().get()
        const tokenParsed = typeof token === 'function' ? `${(await token()).access_token}` : token
        const baseUrlParsed = baseUrl.replace('http', 'ws')

        // Open up the WebSocket connection if it doesn't exist.
        if (!wsInstance) {
          wsInstance = state.setInstance(
            url,
            new WebSocket(`${baseUrlParsed}${endpoint}`, [
              'ttn.lorawan.v3.console.internal.events.v1',
              `ttn.lorawan.v3.header.authorization.bearer.${tokenParsed}`,
            ]),
          )

          // Broadcast connection errors to all listeners.
          wsInstance.addEventListener('error', () => {
            const err = new Error('Error in WebSocket connection')
            const subscriptions = state.getSubscriptions(url)
            for (const s of subscriptions) {
              notify(s[EVENTS.ERROR], err)
              // The error is an error event, but we should only throw proper errors.
              // It has an optional error code that we could use to map to a proper error.
              // However, the error codes are optional and not always used.
              s.reject(err)
            }
          })

          // Event listener for 'close'
          wsInstance.addEventListener('close', closeEvent => {
            // TODO: Handle close event codes.
            // https://github.com/TheThingsNetwork/lorawan-stack/issues/6752
            const subscriptions = state.getSubscriptions(url)
            const wasClean = closeEvent?.wasClean ?? false

            for (const s of subscriptions) {
              notify(s[EVENTS.CLOSE], wasClean)
              if (wasClean) {
                s.resolve()
              } else {
                s.reject(
                  new Error(
                    `WebSocket connection closed unexpectedly with code ${closeEvent.code}`,
                  ),
                )
              }
            }

            state.deleteInstance(url)
          })

          // After the WebSocket connection is open, add the event listeners.
          // Wait for the subscription confirmation message before resolving.
          wsInstance.addEventListener('message', ({ data }) => {
            const dataParsed = JSON.parse(data)
            const sid = dataParsed.id
            const subscription = state.getSubscription(url, sid)

            if (!subscription) {
              warn('Message received for closed or unknown subscription with ID', sid)

              return
            }

            if (dataParsed.type === MESSAGE_TYPES.SUBSCRIBE) {
              notify(subscription[EVENTS.OPEN])
              // Resolve the promise after the subscription confirmation message.
              subscription.resolve()
            }

            if (dataParsed.type === MESSAGE_TYPES.ERROR) {
              notify(subscription[EVENTS.ERROR], dataParsed)
            }

            if (dataParsed.type === MESSAGE_TYPES.PUBLISH) {
              notify(subscription[EVENTS.MESSAGE], dataParsed.event)
            }

            if (dataParsed.type === MESSAGE_TYPES.UNSUBSCRIBE) {
              notify(subscription[EVENTS.CLOSE], subscription.closeRequested)
              // Remove the subscription
              state.deleteSubscription(url, sid)
              if (state.getSubscriptionCount(url) === 0) {
                wsInstance.close()
                state.deleteInstance(url)
              }
            }
          })
        }

        if (wsInstance.readyState === WebSocket.OPEN) {
          // If the WebSocket connection is already open, only add the subscription.
          wsInstance.send(subscriptionPayload)
        } else if (wsInstance.readyState === WebSocket.CONNECTING) {
          // Otherwise wait for the connection to open and then add the subscription.
          const onOpen = () => {
            wsInstance.send(subscriptionPayload)
            wsInstance.removeEventListener('open', onOpen)
          }
          wsInstance.addEventListener('open', onOpen)
        } else {
          reject(new Error('WebSocket connection is closed'))
        }
      } catch (error) {
        const err = error instanceof Error ? error : new Error(error)
        const subscriptions = state.getSubscriptions(url)
        for (const s of subscriptions) {
          notify(s[EVENTS.ERROR], err)
          s.reject(err)
        }
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
      const subscription = state.getSubscription(url, subscriptionId)
      subscription[eventName] = callback

      return this
    },
    close: () => {
      if (
        !wsInstance ||
        wsInstance.readyState === WebSocket.CLOSED ||
        wsInstance.readyState === WebSocket.CLOSING
      ) {
        warn('WebSocket was already closed')
        return Promise.resolve()
      }

      state.markSubscriptionClosing(url, subscriptionId)
      wsInstance.send(unsubscribePayload)

      // Wait for the server to confirm the unsubscribe.
      return new Promise(resolve => {
        const onMessage = ({ data }) => {
          const { type, id } = JSON.parse(data)
          if (id === subscriptionId && type === MESSAGE_TYPES.UNSUBSCRIBE) {
            resolve()
          }
          wsInstance.removeEventListener('message', onMessage)
        }
        wsInstance.addEventListener('message', onMessage)
      })
    },
  }
}
