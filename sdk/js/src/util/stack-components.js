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

/* eslint-disable import/prefer-default-export */

import { STACK_COMPONENTS } from './constants'

/** Takes a list of allowed components and only returns components that have
 * distinct base urls. Used to subscribe to event streaming sources when the
 * stack uses multiple hosts.
 * @param {Array} stackConfig - The stack config object containing base urls per
 * component.
 * @param {Array} components - Components to return distinct ones from.
 * @returns {Array} An array of components that have distinct base urls.
 */
export const getComponentsWithDistinctBaseUrls = function(
  stackConfig,
  components = STACK_COMPONENTS,
) {
  const distinctComponents = components.reduce((collection, component) => {
    if (
      Boolean(stackConfig[component]) &&
      !Object.values(collection).includes(stackConfig[component])
    ) {
      return { ...collection, [component]: stackConfig[component] }
    }
    return collection
  }, {})

  return Object.keys(distinctComponents)
}
