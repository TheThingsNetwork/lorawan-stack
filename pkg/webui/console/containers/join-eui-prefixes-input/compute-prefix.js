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

const BIN_BASE = 2
const HEX_BASE = 16
const CHAR_BYTES = 4

/**
 * Computes the join EUI prefix given `joinEUI` and its `length`.
 * @param {string} joinEUI - The join EUI.
 * @param {number} length - The length of the prefix.
 * @returns {Array} - A list of join EUI prefixes.
 */
function computePrefixes(joinEUI, length = 0) {
  if (length % CHAR_BYTES === 0) {
    return [joinEUI.slice(0, Math.ceil(length / CHAR_BYTES))]
  }

  const charCount = Math.floor(length / CHAR_BYTES)
  const nextCharHex = joinEUI.slice(charCount, charCount + 1)
  const nextCharBinary = parseInt(nextCharHex, HEX_BASE)
    .toString(BIN_BASE)
    .padStart(CHAR_BYTES, '0')

  const rangeStart = parseInt(
    nextCharBinary.slice(0, length % CHAR_BYTES).padEnd(CHAR_BYTES, '0'),
    BIN_BASE,
  )
  const rangeEnd =
    ((rangeStart >>> (CHAR_BYTES - (length % CHAR_BYTES))) + 1) <<
    (CHAR_BYTES - (length % CHAR_BYTES))

  const base = joinEUI.slice(0, charCount)
  const prefixes = []

  for (let i = rangeStart; i < rangeEnd; i++) {
    prefixes.push(base + parseInt(i).toString(HEX_BASE))
  }

  return prefixes
}

export default computePrefixes
