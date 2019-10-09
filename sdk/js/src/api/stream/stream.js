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

import ArrayBufferToString from 'arraybuffer-to-string'
import Token from '../../util/token'
import { notify, EVENTS } from './shared'
import 'web-streams-polyfill/dist/polyfill.js'

/**
 * Opens a new stream.
 *
 * @async
 * @param {Object} payload  - The body of the initial request.
 * @param {string} url - The stream endpoint.
 *
 * @example
 * (async () => {
 *    const stream = await stream(
 *      { identifiers: [{ application_ids: { application_id: 'my-app' }}]},
 *      '/api/v3/events',
 *    )
 *
 *    // add listeners to the stream
 *    stream
 *      .on('start', () => console.log('conn opened'));
 *      .on('chunk', chunk => console.log('received chunk', chunk));
 *      .on('error', error => console.log(error));
 *      .on('close', () => console.log('conn closed'))
 *
 *    // close the stream after 20 s
 *    setTimeout(() => stream.close(), 20000)
 * })()
 *
 * @returns {Object} The stream subscription object with the `on` function for
 * attaching listeners and the `close` function to close the stream.
 */
export default async function(payload, url) {
  let listeners = Object.values(EVENTS).reduce((acc, curr) => ({ ...acc, [curr]: null }), {})
  const token = new Token().get()

  let Authorization = null
  if (typeof token === 'function') {
    Authorization = `Bearer ${(await token()).access_token}`
  } else {
    Authorization = `Bearer ${token}`
  }

  let reader = null
  const response = await fetch(url, {
    body: JSON.stringify(payload),
    method: 'POST',
    headers: {
      Authorization,
    },
  })

  if (response.status !== 200) {
    const err = await response.json()

    throw 'error' in err ? err.error : err
  }

  reader = response.body.getReader()
  reader
    .read()
    .then(function(data) {
      notify(listeners[EVENTS.START])

      return data
    })
    .then(function onChunk({ done, value }) {
      if (done) {
        notify(listeners[EVENTS.CLOSE])
        listeners = null
        return
      }

      const parsed = ArrayBufferToString(value)

      for (const line of parsed.trim().split('\n')) {
        const result = JSON.parse(line).result
        notify(listeners[EVENTS.CHUNK], result)
      }

      return reader.read().then(onChunk)
    })
    .catch(function(error) {
      notify(listeners[EVENTS.ERROR], error)
      listeners = null
    })

  return {
    on(eventName, callback) {
      if (listeners[eventName] === undefined) {
        throw new Error(
          `${eventName} event is not supported. Should be one of: start, error, chunk or close`,
        )
      }

      listeners[eventName] = callback

      return this
    },
    close() {
      if (reader) {
        reader.cancel()
      }
    },
  }
}
