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

/**
 * Generates a random string of the given length.
 * @param {number} len - The length of the string to generate.
 * @param {string} [type='hex'] - The type of the string to generate.
 * @returns {string} The generated string.
 * @example
 * const randomString = randomBytes(16)
 * @example
 * const randomString = randomBytes(16, 'base64')
 */
export default (len, type = 'hex') => {
  let bytes
  if (window.crypto) {
    bytes = crypto.getRandomValues(new Uint8Array(Math.floor(len / 2)))
  } else {
    const byteLength = Math.floor(len / 2)
    bytes = new Uint8Array(byteLength)
    for (let i = 0; i < byteLength; i++) {
      bytes[i] = Math.floor(Math.random() * 256)
    }
  }
  switch (type) {
    case 'base64':
      return btoa(String.fromCharCode(...bytes))
    case 'hex':
    default:
      return Array.from(bytes, byte => byte.toString(16).padStart(2, '0'))
        .join('')
        .toUpperCase()
  }
}
