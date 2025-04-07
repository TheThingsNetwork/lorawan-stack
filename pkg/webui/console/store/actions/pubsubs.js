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

import createRequestActions from '@ttn-lw/lib/store/actions/create-request-actions'
import {
  createPaginationBaseActionType,
  createPaginationByIdRequestActions,
} from '@ttn-lw/lib/store/actions/pagination'

export const CREATE_PUBSUB_BASE = 'CREATE_PUBSUB'
export const [
  { request: CREATE_PUBSUB, success: CREATE_PUBSUB_PUBSUB, failure: CREATE_PUBSUB_FAILURE },
  { request: createPubsub, success: createPubsubSuccess, failure: createPubsubFailure },
] = createRequestActions(CREATE_PUBSUB_BASE, (appId, pubsub) => ({ appId, pubsub }))

export const GET_PUBSUB_BASE = 'GET_PUBSUB'
export const [
  { request: GET_PUBSUB, success: GET_PUBSUB_SUCCESS, failure: GET_PUBSUB_FAILURE },
  { request: getPubsub, success: getPubsubSuccess, failure: getPubsubFailure },
] = createRequestActions(
  GET_PUBSUB_BASE,
  (appId, pubsubId) => ({ appId, pubsubId }),
  (appId, pubsubId, selector) => ({ selector }),
)

export const GET_PUBSUBS_LIST_BASE = createPaginationBaseActionType('PUBSUB')
export const [
  {
    request: GET_PUBSUBS_LIST,
    success: GET_PUBSUBS_LIST_SUCCESS,
    failure: GET_PUBSUBS_LIST_FAILURE,
  },
  { request: getPubsubsList, success: getPubsubsListSuccess, failure: getPubsubsListFailure },
] = createPaginationByIdRequestActions('PUBSUB')

export const UPDATE_PUBSUB_BASE = 'UPDATE_PUBSUB'
export const [
  { request: UPDATE_PUBSUB, success: UPDATE_PUBSUB_SUCCESS, failure: UPDATE_PUBSUB_FAILURE },
  { request: updatePubsub, success: updatePubsubSuccess, failure: updatePubsubFailure },
] = createRequestActions(UPDATE_PUBSUB_BASE, (appId, pubsubId, patch) => ({
  appId,
  pubsubId,
  patch,
}))

export const DELETE_PUBSUB_BASE = 'DELETE_PUBSUB'
export const [
  { request: DELETE_PUBSUB, success: DELETE_PUBSUB_SUCCESS, failure: DELETE_PUBSUB_FAILURE },
  { request: deletePubsub, success: deletePubsubSuccess, failure: deletePubsubFailure },
] = createRequestActions(DELETE_PUBSUB_BASE, (appId, pubsubId) => ({ appId, pubsubId }))
