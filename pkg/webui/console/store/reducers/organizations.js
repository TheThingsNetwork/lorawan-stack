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

import { getOrganizationId } from '../../../lib/selectors/id'
import { GET_ORGS_LIST_SUCCESS, CREATE_ORG_SUCCESS } from '../actions/organizations'

const organization = function(state = {}, organization) {
  return {
    ...state,
    ...organization,
  }
}

const defaultState = {
  entities: {},
}

const organizations = function(state = defaultState, { type, payload }) {
  switch (type) {
    case GET_ORGS_LIST_SUCCESS:
      const entities = payload.entities.reduce(
        function(acc, org) {
          const id = getOrganizationId(org)

          acc[id] = organization(acc[id], org)
          return acc
        },
        { ...state.entities },
      )

      return {
        ...state,
        entities,
      }
    case CREATE_ORG_SUCCESS:
      const id = getOrganizationId(payload)

      return {
        ...state,
        [id]: organization(state[id], payload),
      }
    default:
      return state
  }
}

export default organizations
