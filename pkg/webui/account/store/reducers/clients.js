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

import { EVENT_END_DEVICE_HEARTBEAT_FILTERS_REGEXP } from '@console/constants/event-filters'

import getByPath from '@ttn-lw/lib/get-by-path'

import {
  GET_CLIENT,
  GET_CLIENT_SUCCESS,
  GET_CLIENTS_LIST,
  GET_CLIENTS_LIST_SUCCESS,
  UPDATE_CLIENT_SUCCESS,
  DELETE_CLIENT_SUCCESS,
} from '@account/store/actions/clients'

const defaultState = {
  fetching: false,
  clients_list: undefined,
  totalCount: null,
  selectedClient: null,
  error: false,
}

const clients = (state = defaultState, { type, payload }) => {
  switch (type) {
    case GET_CLIENT:
      return {
        ...state,
        selectedClient: payload.id,
      }
    case GET_CLIENTS_LIST:
      return {
        ...state,
        fetching: true,
        clients_list: undefined,
        error: false,
      }
    case GET_CLIENTS_LIST_SUCCESS:
      return {
        ...state,
        fetching: false,
        clients_list: [...payload.entities],
        totalCount: payload.totalCount,
        error: false,
      }
    default:
      return state
  }
}

export default clients
