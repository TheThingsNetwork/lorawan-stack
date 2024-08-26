// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

import { combineActions, handleActions } from 'redux-actions'

import { APPLICATION, END_DEVICE, GATEWAY } from '@console/constants/entities'

import { isNotFoundError, isPermissionDeniedError } from '@ttn-lw/lib/errors/utils'
import yup from '@ttn-lw/lib/yup'

import {
  TRACK_RECENCY_FREQUENCY_ITEM,
  DELETE_RECENCY_FREQUENCY_ITEM,
} from '@console/store/actions/recency-frequency-items'

import { DELETE_APP_SUCCESS, GET_APP_FAILURE } from '../actions/applications'
import { DELETE_GTW_SUCCESS, GET_GTW_FAILURE } from '../actions/gateways'
import { DELETE_DEV_SUCCESS, GET_DEVICES_LIST_FAILURE, GET_DEV_FAILURE } from '../actions/devices'
import { APPLY_PERSISTED_STATE_SUCCESS } from '../actions/user'

const KEY_PATTERN = /^(END_DEVICE|APPLICATION|GATEWAY):[a-z0-9](?:[-]?[a-z0-9/]){2,}$/

const defaultState = {
  items: {},
}

const removeItem = (state, type, id) => {
  const { [`${type}:${id}`]: _, ...newState } = state.items

  return {
    ...state,
    items: newState,
  }
}

const schema = yup
  .object({
    frequency: yup.number().required(),
    lastAccessed: yup.number().required(),
  })
  .noUnknown(true)

export const validateRecencyFrequencyEntities = entities => {
  const validEntities = Object.entries(entities.items).reduce((acc, [key, value]) => {
    if (!KEY_PATTERN.test(key)) {
      throw new Error('Invalid key pattern')
    }
    schema.validateSync(value)
    acc[key] = value
    return acc
  }, {})
  return { items: validEntities }
}

export default handleActions(
  {
    [APPLY_PERSISTED_STATE_SUCCESS]: (state, { payload }) => {
      if (payload.recencyFrequencyItems) {
        return validateRecencyFrequencyEntities(payload.recencyFrequencyItems)
      }

      return state
    },
    [TRACK_RECENCY_FREQUENCY_ITEM]: (state, { payload: { type, id } }) => {
      const entityKey = `${type}:${id}`
      const entity = state.items[entityKey]

      const newEntity = entity
        ? {
            frequency: entity.frequency + 1,
            lastAccessed: Date.now(),
          }
        : {
            frequency: 1,
            lastAccessed: Date.now(),
          }

      return {
        ...state,
        items: {
          ...state.items,
          [entityKey]: newEntity,
        },
      }
    },
    [DELETE_RECENCY_FREQUENCY_ITEM]: (state, { payload }) => {
      const { type, id } = payload

      return removeItem(state, type, id)
    },
    [combineActions(GET_APP_FAILURE, GET_DEVICES_LIST_FAILURE)]: (state, { payload, meta }) => {
      if (isNotFoundError(payload)) {
        return removeItem(state, APPLICATION, meta.requestPayload.id)
      } else if (isPermissionDeniedError(payload)) {
        // Permission denied is not necessarily conclusive that the entity cannot be
        // accessed at all, since it's possible that only certain parts of the entity
        // are not accessible. For now, we will remove the entity either way but
        // this could be improved in the future.
        return removeItem(state, APPLICATION, meta.requestPayload.id)
      }

      return state
    },
    [GET_GTW_FAILURE]: (state, { payload, meta }) => {
      if (isNotFoundError(payload)) {
        return removeItem(state, GATEWAY, meta.requestPayload.id)
      } else if (isPermissionDeniedError(payload)) {
        // See above.
        return removeItem(state, GATEWAY, meta.requestPayload.id)
      }

      return state
    },
    [GET_DEV_FAILURE]: (state, { payload, meta }) => {
      const id = `${meta.requestPayload.appId}/${meta.requestPayload.deviceId}`
      if (isNotFoundError(payload)) {
        return removeItem(state, END_DEVICE, id)
      } else if (isPermissionDeniedError(payload)) {
        // See above.
        return removeItem(state, END_DEVICE, id)
      }

      return state
    },
    [DELETE_APP_SUCCESS]: (state, { payload }) => removeItem(state, APPLICATION, payload.id),
    [DELETE_GTW_SUCCESS]: (state, { payload }) => removeItem(state, GATEWAY, payload.id),
    [DELETE_DEV_SUCCESS]: (state, { payload }) => removeItem(state, END_DEVICE, payload.id),
  },
  defaultState,
)
