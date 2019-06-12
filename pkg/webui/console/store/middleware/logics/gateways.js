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

import api from '../../../api'
import * as gateways from '../../actions/gateways'
import createRequestLogic from './lib'

const getGatewaysLogic = createRequestLogic({
  type: gateways.GET_GTWS_LIST,
  latest: true,
  async process ({ action }) {
    const { page, pageSize: limit, query } = action.payload.filters
    const data = query
      ? await api.gateways.search({
        page,
        limit,
        id_contains: query,
        name_contains: query,
      })
      : await api.gateways.list({ page, limit }, [ 'name,description,frequency_plan_id' ])
    return {
      gateways: data.gateways,
      totalCount: data.totalCount,
    }
  },
})

const getGatewaysRightsLogic = createRequestLogic({
  type: gateways.GET_GTWS_RIGHTS_LIST,
  async process ({ action }, dispatch, done) {
    const { id } = action.payload
    const result = await api.rights.gateways(id)
    return result.rights.sort()
  },
})

export default [
  getGatewaysLogic,
  getGatewaysRightsLogic,
]
