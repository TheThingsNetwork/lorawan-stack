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

import axios from 'axios'
import Token from '../../util/token'
import { notify, EVENTS } from './shared'

/**
 * Opens a new stream.
 *
 * @param {Object} payload  - The body of the initial request.
 * @param {string} url - The stream endpoint.
 *
 * @example
 * (async () => {
 *    const stream = await stream(
 *      { identifiers: [{ application_ids: { application_id: 'my-app' }}]},
 *      'http://localhost:1885/api/v3/events',
 *    )
 *
 *    // add listeners to the stream
 *    stream
 *      .on('start', () => console.log('conn opened'));
 *      .on('event', message => console.log('received event message', message));
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
export default async function (payload, url) {
  let listeners = Object.values(EVENTS)
    .reduce((acc, curr) => ({ ...acc, [curr]: null }), {})
  const token = new Token().get()

  let Authorization = null
  if (typeof token === 'function') {
    Authorization = `Bearer ${(await token()).access_token}`
  } else {
    Authorization = `Bearer ${token}`
  }

  let reader = null
  axios({
    url,
    data: JSON.stringify(payload),
    method: 'POST',
    responseType: 'stream',
    headers: {
      Authorization,
    },
  })
    .then(response => response.data)
    .then(function (stream) {
      reader = stream
      notify(listeners[EVENTS.START])

      stream.on('data', function (data) {
        const parsed = data.toString('utf8')
        const result = JSON.parse(parsed).result
        notify(listeners[EVENTS.EVENT], result)
      })
      stream.on('end', function () {
        notify(listeners[EVENTS.CLOSE])
        listeners = null
      })
      stream.on('error', function (error) {
        notify(listeners[EVENTS.ERROR], error)
        listeners = null
      })
    })

  return {
    on (eventName, callback) {
      if (listeners[eventName] === undefined) {
        throw new Error(
          `${eventName} event is not supported. Should be one of: start, error, event or close`
        )
      }

      listeners[eventName] = callback

      return this
    },
    close () {
      if (reader) {
        reader.cancel()
      }
    },
  }
}
