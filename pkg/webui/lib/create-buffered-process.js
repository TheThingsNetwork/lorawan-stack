// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
 * `createBufferedProcess` creates a buffered event ingest to combine and
 * process multiple payloads at once that can otherwise cause lags when
 * processed individually in rapid succession.
 *
 * @param {Function} process - The action to be performed on the combined
 * payloads before the buffer is cleared.
 * @param {number} maxItems - The maximum number of items that the buffer can
 * hold before it will dispatch, regardless of the delay. Default is 20.
 * @param {number} delay - The delay in ms that the buffer will wait to collect
 * more items after receiving the last message. Default is 200ms.
 * @returns {object} An object containing the `addToBuffer` handler to push
 * items into the buffer and the `clearBuffer` hook to trigger a processing and
 * clearing of the buffer manually.
 */
const createBufferedProcess = (process, maxItems = 20, delay = 200) => {
  let buffer = []
  let timer

  const commit = () => {
    process(buffer)
    buffer = []
    if (timer) {
      clearTimeout(timer)
    }
  }

  return {
    addToBuffer: msg => {
      buffer.push(msg)
      if (buffer.length > maxItems) {
        commit()
      } else {
        clearTimeout(timer)
        timer = setTimeout(commit, delay)
      }
    },
    clearBuffer: commit,
  }
}

export default createBufferedProcess
