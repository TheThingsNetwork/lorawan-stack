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
import { APPLICATION, END_DEVICE, GATEWAY, ORGANIZATION } from '@console/constants/entities'

import createRequestLogic from '@ttn-lw/lib/store/logics/create-request-logic'
import {
  getApplicationId,
  getDeviceId,
  getGatewayId,
  getOrganizationId,
} from '@ttn-lw/lib/selectors/id'

import * as search from '@console/store/actions/search'

import { selectConcatenatedTopEntitiesByType } from '@console/store/selectors/top-entities'

const getGlobalSearchResults = createRequestLogic({
  type: search.GET_GLOBAL_SEARCH_RESULTS,
  process: async ({ getState, action }) => {
    const { query } = action.payload
    const params = {
      page: 1,
      limit: 10,
      query,
      order: undefined,
      deleted: false,
    }

    const topApplications = selectConcatenatedTopEntitiesByType(getState(), APPLICATION).slice(0, 3)

    const responses = await Promise.all([
      tts.Applications.search(params, ['name']),
      Promise.all(
        topApplications.map(app => tts.Applications.Devices.search(app.id, params, ['name'])),
      ),
      tts.Gateways.search(params, ['name']),
      tts.Organizations.search(params, ['name']),
    ])

    const results = [
      {
        category: APPLICATION,
        items: responses[0].applications.map(app => ({
          id: getApplicationId(app),
          type: APPLICATION,
          path: `/applications/${getApplicationId(app)}`,
          ...app,
        })),
        totalCount: responses[0].totalCount,
      },
      {
        category: END_DEVICE,
        items: responses[1]
          // Combine all end devices from all applications together
          .reduce(
            (acc, res) => {
              acc.end_devices = acc.end_devices.concat(res.end_devices)
              acc.totalCount += res.totalCount
              return acc
            },
            { end_devices: [], totalCount: 0 },
          )
          .end_devices.map(device => ({
            id: getDeviceId(device),
            type: END_DEVICE,
            path: `/applications/${getApplicationId(device)}/devices/${getDeviceId(device)}`,
            ...device,
          })),
        totalCount: responses[1].totalCount,
      },
      {
        category: GATEWAY,
        items: responses[2].gateways.map(gateway => ({
          id: getGatewayId(gateway),
          type: GATEWAY,
          path: `/gateways/${getGatewayId(gateway)}`,
          ...gateway,
        })),
        totalCount: responses[2].totalCount,
      },
      {
        category: ORGANIZATION,
        items: responses[3].organizations.map(org => ({
          id: getOrganizationId(org),
          type: ORGANIZATION,
          path: `/organizations/${getOrganizationId(org)}`,
          ...org,
        })),
        totalCount: responses[3].totalCount,
      },
    ]

    return {
      query,
      results,
    }
  },
})

export default [getGlobalSearchResults]
