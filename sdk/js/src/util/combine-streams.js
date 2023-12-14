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

/**
 * Combines multiple streams into a single subscription provider.
 *
 * @param {Array} streams - An array of (async) stream functions.
 * @returns {object} The stream subscription object with the `close` function
 * to close the stream.
 */
const combinedStream = async streams => {
  if (!(streams instanceof Array) || streams.length === 0) {
    throw new Error('Cannot combine streams with invalid stream array.')
  } else if (streams.length === 1) {
    return streams[0]
  }

  try {
    const subscribers = await Promise.all(streams)
    return {
      close: () => Promise.all(subscribers.map(subscriber => subscriber.close())),
    }
  } catch (error) {
    // Ensure that if only some streams fail, the successful ones are closed.
    await Promise.all(
      streams.map(async stream => {
        try {
          const subscriber = await stream
          await subscriber.close()
        } catch {}
      }),
    )
    throw error
  }
}

export default combinedStream
