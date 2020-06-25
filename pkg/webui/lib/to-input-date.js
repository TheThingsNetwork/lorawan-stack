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
 * Function to return a valid input date string from a Date object.
 *
 * @param {object} d - Date() object.
 * @returns {string} 'yyyy-mm-dd' or undefined.
 */

export default function(d) {
  if (Object.prototype.toString.call(d) !== '[object Date]' || isNaN(d.getTime())) {
    return undefined
  }
  const mm = d.getMonth() + 1
  const dd = d.getDate()
  const yy = d.getFullYear()
  return `${yy}-${`0${mm}`.slice(-2)}-${`0${dd}`.slice(-2)}`
}
