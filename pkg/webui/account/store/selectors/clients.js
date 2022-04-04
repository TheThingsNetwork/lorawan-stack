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

import { createFetchingSelector } from '@ttn-lw/lib/store/selectors/fetching'
import { createErrorSelector } from '@ttn-lw/lib/store/selectors/error'

import { GET_CLIENTS_LIST_BASE } from '@account/store/actions/clients'

const selectClientsStore = state => state.clients

export const selectClientById = (state, id) => selectClientsStore(state)[id]

const selectClientsFetching = createFetchingSelector(GET_CLIENTS_LIST_BASE)
const selectClientsError = createErrorSelector(GET_CLIENTS_LIST_BASE)

export const selectOAuthClients = state => selectClientsStore(state).clients_list
export const selectOAuthClientsTotalCount = state => selectClientsStore(state).totalCount
export const selectOAuthClientsFetching = state => selectClientsFetching(state)
export const selectOAuthClientsError = state => selectClientsError(state)
