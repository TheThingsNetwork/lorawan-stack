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

import { GET_PUBSUB_BASE, GET_PUBSUBS_LIST_BASE } from '../actions/pubsubs'
import { createFetchingSelector } from './fetching'
import { createErrorSelector } from './error'

const selectPubsubStore = state => state.pubsubs

// Pubsub
export const selectPubsubEntityStore = state => selectPubsubStore(state).entities
export const selectSelectedPubsubId = state => selectPubsubStore(state).selectedPubsub
export const selectSelectedPubsub = state =>
  selectPubsubEntityStore(state)[selectSelectedPubsubId(state)]
export const selectPubsubError = createErrorSelector(GET_PUBSUB_BASE)
export const selectPubsubFetching = createFetchingSelector(GET_PUBSUB_BASE)

// Pubsubs
export const selectPubsubs = state => Object.values(selectPubsubEntityStore(state))
export const selectPubsubsTotalCount = state => selectPubsubEntityStore(state).totalCount
export const selectPubsubsFetching = createFetchingSelector(GET_PUBSUBS_LIST_BASE)
export const selectPubsubsError = createErrorSelector(GET_PUBSUBS_LIST_BASE)
