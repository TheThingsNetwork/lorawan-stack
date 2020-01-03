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

import { GET_PUBSUB, GET_PUBSUB_SUCCESS, GET_PUBSUBS_LIST_SUCCESS } from '../actions/pubsubs'
import { getPubsubId } from '../../../lib/selectors/id'

const defaultState = {
  selectedPubsub: null,
  totalCount: undefined,
  entities: {},
}

const pubsubs = function(state = defaultState, { type, payload }) {
  switch (type) {
    case GET_PUBSUB:
      return {
        ...state,
        selectedPubsub: payload.pubsubId,
      }
    case GET_PUBSUB_SUCCESS:
      return {
        ...state,
        entities: {
          ...state.entities,
          [getPubsubId(payload)]: payload,
        },
      }
    case GET_PUBSUBS_LIST_SUCCESS:
      return {
        ...state,
        entities: {
          ...payload.entities.reduce((acc, pubsub) => {
            acc[getPubsubId(pubsub)] = pubsub
            return acc
          }, {}),
        },
        totalCount: payload.totalCount,
      }
    default:
      return state
  }
}

export default pubsubs
