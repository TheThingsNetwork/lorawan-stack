// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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
 * Takes a templated string and interpolates its values using a values object.
 *
 * @param {string} str - The to be interpolated template string.
 * @param {values} values - The values to interpolate the template string with.
 *
 * @returns {string} - The interpolated string.
 */
const interpolate = (str, values = {}) => str.replace(/\{([^}]+)\}/g, (a, b) => values[b])

export default interpolate
