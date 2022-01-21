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
 * @example
 * const props = {
 *  className: 'objClass',
 *  counter: 3,
 *  'data-test-id': 'object-test',
 * }
 *
 * const testOnly = filterDataProps(props) // result is {'data-test-id': 'object-test'}
 *
 * @param {object} props - Multinested object.
 * @returns {object} New object with input object properties that start with `data`.
 */
export default props =>
  Object.keys(props)
    .filter(key => key.startsWith('data-'))
    .reduce((acc, key) => {
      acc[key] = props[key]
      return acc
    }, {})
