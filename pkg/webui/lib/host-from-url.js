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
 * Extracts hostname from the `url`.
 *
 * @param {string} url - The URL string.
 * @returns {string?} - The hostname of the `url` or undefined.
 */
export default url => {
  try {
    if (url.match(/^[a-zA-Z0-9]+:\/\/.*/)) {
      return new URL(url).hostname
    }

    return new URL(`http://${url}`).hostname
  } catch (error) {
    return url
  }
}
