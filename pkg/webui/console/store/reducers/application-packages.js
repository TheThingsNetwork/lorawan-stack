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

import { isUndefined, omitBy } from 'lodash'

import {
  GET_APP_PKG_DEFAULT_ASSOC_SUCCESS,
  SET_APP_PKG_DEFAULT_ASSOC_SUCCESS,
  DELETE_APP_PKG_DEFAULT_ASSOC_SUCCESS,
} from '@console/store/actions/application-packages'

const defaultState = {
  default: {},
}

const applicationPackages = (state = defaultState, { type, payload }) => {
  switch (type) {
    case GET_APP_PKG_DEFAULT_ASSOC_SUCCESS:
    case SET_APP_PKG_DEFAULT_ASSOC_SUCCESS:
      return {
        ...state,
        default: omitBy(
          {
            ...state.default,
            [payload.ids.f_port]: 'created_at' in payload ? payload : undefined,
          },
          isUndefined,
        ),
      }
    case DELETE_APP_PKG_DEFAULT_ASSOC_SUCCESS:
      const { [payload.fPort]: deleted, ...rest } = state.default

      return {
        ...state,
        default: rest,
      }
    default:
      return state
  }
}

export default applicationPackages
