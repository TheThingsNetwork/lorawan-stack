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

const newSubscription = (resolve, reject, unsubscribe) => {
  const listeners = { ...initialListeners }
  let closeRequested = false
  const externalSubscription = {
    on: (eventName, callback) => {
      if (!Object.values(EVENTS).includes(eventName)) {
        throw new Error(
          `${eventName} event is not supported. Should be one of: open, message, error or close`,
        )
      }
      listeners[eventName] = callback
      return externalSubscription
    },
    close: () => {
      closeRequested = true
      return unsubscribe()
    },
  }
  return {
    onError: err => {
      notify(listeners[EVENTS.ERROR], err)
      // If an error occurs while we are trying to subscribe, we should reject
      // the promise in order to propagate the implicit subscription failure.
      reject(err)
    },
    onClose: closeEvent => {
      notify(listeners[EVENTS.CLOSE], closeRequested)
      // If the connection has been closed while we are trying subscribe, we should
      // reject the promise in order to propagate the implicit subscription failure.
      reject(new Error(`WebSocket connection closed unexpectedly with code ${closeEvent.code}`))
    },
    onMessage: dataParsed => {
      if (dataParsed.type === MESSAGE_TYPES.SUBSCRIBE) {
        notify(listeners[EVENTS.OPEN])
        // Resolve the promise after the subscription confirmation message.
        resolve(externalSubscription)
      }

      if (dataParsed.type === MESSAGE_TYPES.ERROR) {
        notify(listeners[EVENTS.ERROR], dataParsed.error)
      }

      if (dataParsed.type === MESSAGE_TYPES.PUBLISH) {
        notify(listeners[EVENTS.MESSAGE], dataParsed.event)
      }

      if (dataParsed.type === MESSAGE_TYPES.UNSUBSCRIBE) {
        notify(listeners[EVENTS.CLOSE], closeRequested)
      }
    },
  }
}

const newInstance = (wsInstance, onClose) => {
  const subscriptions = {}
  let closeRequested = false

  // Broadcast connection errors to all subscriptions.
  wsInstance.addEventListener('error', () => {
    const err = new Error('Error in WebSocket connection')
    for (const subscription of Object.values(subscriptions)) {
      subscription.onError(err)
    }
  })

  // Broadcast connection closure to all subscriptions.
  wsInstance.addEventListener('close', closeEvent => {
    if (closeRequested) {
      // If the close has been requested already, the instance has been
      // deregistered and there are no subscriptions left.
      return
    }
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
      warn('Message received for closed or unknown subscription with ID', sid)
      return
    }

    subscription.onMessage(dataParsed)

    if (dataParsed.type === MESSAGE_TYPES.UNSUBSCRIBE) {
      delete subscriptions[sid]
      if (Object.keys(subscriptions).length === 0) {
        closeRequested = true
        wsInstance.close()
        onClose()
      }
    }
  })

  return {
    subscribe: (sid, resolve, reject, subscribePayload, unsubscribePayload) => {
      if (sid in subscriptions) {
        throw new Error(`Subscription with ID ${sid} already exists`)
      }

      let unsubscribed = null
      const unsubscribe = () => {
        if (unsubscribed) {
          return unsubscribed
        }

        if (
          wsInstance.readyState === WebSocket.CLOSED ||
          wsInstance.readyState === WebSocket.CLOSING
        ) {
          warn('WebSocket was already closed')
          return Promise.resolve()
        }

        wsInstance.send(unsubscribePayload)

        // Wait for the server to confirm the unsubscribe.
        unsubscribed = new Promise(resolve => {
          const onMessage = ({ data }) => {
            const { type, id } = JSON.parse(data)
            if (id === sid && type === MESSAGE_TYPES.UNSUBSCRIBE) {
              resolve()
              wsInstance.removeEventListener('message', onMessage)
            }
          }
          wsInstance.addEventListener('message', onMessage)
          const onClose = () => {
            resolve()
            wsInstance.removeEventListener('close', onClose)
          }
          wsInstance.addEventListener('close', onClose)
        })
        return unsubscribed
      }

      const subscription = newSubscription(resolve, reject, unsubscribe)
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
  const token = new Token().get()
  const tokenParsed = typeof token === 'function' ? `${(await token()).access_token}` : token
  const baseUrlParsed = baseUrl.replace('http', 'ws')

  return await Promise.race([
    new Promise((resolve, reject) => {
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
      instance.subscribe(subscriptionId, resolve, reject, subscriptionPayload, unsubscribePayload)
    }),
    new Promise((_resolve, reject) => setTimeout(() => reject(new Error('timeout')), timeout)),
  ])
}
