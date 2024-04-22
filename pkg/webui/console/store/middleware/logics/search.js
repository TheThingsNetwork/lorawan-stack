// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

import tts from '@console/api/tts'

import createRequestLogic from '@ttn-lw/lib/store/logics/create-request-logic'
import { getApplicationId, getGatewayId, getOrganizationId } from '@ttn-lw/lib/selectors/id'

import * as search from '@console/store/actions/search'

const getGlobalSearchResults = createRequestLogic({
  type: search.GET_GLOBAL_SEARCH_RESULTS,
  process: async ({ action }) => {
    const { query } = action.payload
    const params = {
      page: 1,
      limit: 10,
      query,
      order: undefined,
      deleted: false,
    }

    const responses = await Promise.all([
      tts.Applications.search(params, ['name']),
      tts.Gateways.search(params, ['name']),
      tts.Organizations.search(params, ['name']),
    ])

    const results = [
      {
        category: 'applications',
        items: responses[0].applications.map(app => ({
          id: getApplicationId(app),
          path: `/applications/${getApplicationId(app)}`,
          ...app,
        })),
        totalCount: responses[0].totalCount,
      },
      {
        category: 'gateways',
        items: responses[1].gateways.map(gateway => ({
          id: getGatewayId(gateway),
          path: `/gateways/${getGatewayId(gateway)}`,
          ...gateway,
        })),
        totalCount: responses[1].totalCount,
      },
      {
        category: 'organizations',
        items: responses[2].organizations.map(org => ({
          id: getOrganizationId(org),
          path: `/organizations/${getOrganizationId(org)}`,
          ...org,
        })),
        totalCount: responses[2].totalCount,
      },
    ]

    return {
      query,
      results,
    }
  },
})

export default [getGlobalSearchResults]
