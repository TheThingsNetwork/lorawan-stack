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

import { EVENTS, INITIAL_LISTENERS, notify } from './shared'
import subscribeToWebSocketStream from './subscribeToWebSocketStream'

/*
 * Subscribe to an event stream with multiple base URLs.
 * Semantically equivalent to subscribeToWebSocketStream, but guarantees uniform stream closure and unaffected
 * event emission in case of failure of one or more streams.
 */
export default async (
  payload,
  baseUrls,
  listeners,
  endpoint = '/console/internal/events/',
  timeout = 10000,
) => {
  if (!(baseUrls instanceof Array) || baseUrls.length === 0) {
    throw new Error('Cannot subscribe to events without base URLs')
  }
  if (baseUrls.length === 1) {
    return subscribeToWebSocketStream(payload, baseUrls[0], listeners, endpoint, timeout)
  }

  for (const eventName of Object.keys(listeners)) {
    if (!Object.values(EVENTS).includes(eventName)) {
      throw new Error(
        `${eventName} event is not supported. Should be one of: message, error or close`,
      )
    }
  }
  const filledListeners = { ...INITIAL_LISTENERS, ...listeners }

  // Interweaving multiple streams has the side effect of making the standard event listener
  // state machine look erratic externally - only certain streams may fail, while others may continue
  // to work. Upper layers which use the stream must not be exposed to this detail - once one
  // stream fails, we need to ensure that all streams are closed and no stray events are emitted.
  //
  // The standard state machine guarantees that once an error or close event is emitted, no further
  // message events will be emitted. It is also guaranteed that once an error event is emitted, a close
  // event will follow.
  let [closeAll, hadError, hadClose] = [() => {}, false, false]
  const uniformListeners = {
    [EVENTS.MESSAGE]: (...params) => {
      if (hadClose || hadError) {
        return
      }
      notify(filledListeners[EVENTS.MESSAGE], ...params)
    },
    [EVENTS.ERROR]: (...params) => {
      if (hadClose || hadError) {
        return
      }
      hadError = true
      notify(filledListeners[EVENTS.ERROR], ...params)
    },
    [EVENTS.CLOSE]: (...params) => {
      if (hadClose) {
        return
      }
      hadClose = true
      notify(filledListeners[EVENTS.CLOSE], ...params)
      closeAll()
    },
  }

  const pendingStreams = baseUrls.map(baseUrl =>
    subscribeToWebSocketStream(payload, baseUrl, uniformListeners, endpoint, timeout),
  )
  try {
    const streams = await Promise.all(pendingStreams)
    closeAll = () => Promise.all(streams.map(stream => stream.close()))
    return {
      open: () => streams.forEach(stream => stream.open()),
      close: closeAll,
    }
  } catch (error) {
    // Ensure that if only some streams fail, the successful ones are closed.
    await Promise.all(
      pendingStreams.map(async pendingStream => {
        try {
          const stream = await pendingStream
          await stream.close()
        } catch {
          // Only the pending stream promise may throw, as `close` does not throw.
          // Although multiple streams may fail, we will rethrow only the first error
          // and ignore the rest, as they are not really actionable.
        }
      }),
    )
    throw error
  }
}
