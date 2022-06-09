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

import { isPlainObject } from 'lodash'

/**
 * Combines multiple refs.
 *
 * @param {Array<object>} refs - A list of refs to be merged.
 * @returns {Function} - The ref callback with the DOM element that is assigned to every ref in `refs`.
 */
const combineRefs = refs => val => {
  refs.forEach(ref => {
    if (isPlainObject(ref)) {
      ref.current = val
    } else if (typeof ref === 'function') {
      ref(val)
    }
  })
}

export default combineRefs
