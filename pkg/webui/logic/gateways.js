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

import { createLogic } from 'redux-logic'

import api from '../api'
import * as gateways from '../actions/gateways'

const getGatewaysLogic = createLogic({
  type: [
    gateways.GET_GTWS_LIST,
    gateways.CHANGE_GTWS_ORDER,
    gateways.CHANGE_GTWS_PAGE,
    gateways.SEARCH_GTWS_LIST,
  ],
  latest: true,
  async process ({ getState, action }, dispatch, done) {
    const { filters } = action

    try {
      const data = filters.query
        ? await api.v3.is.gateways.search(filters)
        : await api.v3.is.gateways.list(filters)
      const gtws = data.gateways.map(g => ({ ...g, antennasCount: g.antennas.length }))
      dispatch(gateways.getGatewaysSuccess(gtws, data.totalCount))
    } catch (error) {
      dispatch(gateways.getGatewaysFailure(error))
    }

    done()
  },
})

export default [
  getGatewaysLogic,
]
