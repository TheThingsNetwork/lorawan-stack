// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

const client = function (state = [], action) {
  switch (action.type) {
  case 'GET_CLIENT':
    return {
      ...state,
      [action.clientId]: { fetching: true },
    }
  case 'GET_CLIENT_SUCCESS':
    const clientId = action.clientData.ids.client_id
    return {
      ...state,
      [clientId]: {
        client: action.clientData,
        fetching: false,
      },
    }
  case 'GET_CLIENT_FAILURE':
    return {
      ...state,
      [action.clientId]: {
        client: null,
        fetching: false,
      },
    }
  default:
    return state
  }
}

export default client
