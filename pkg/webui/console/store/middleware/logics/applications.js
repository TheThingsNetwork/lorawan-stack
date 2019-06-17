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
import * as applications from '../../actions/applications'
import createRequestLogic from './lib'

const getApplicationsLogic = createRequestLogic({
  type: applications.GET_APPS_LIST,
  latest: true,
  async process ({ action }) {
    const { payload: { page, limit, query }} = action

    const data = query
      ? await api.applications.search({
        page,
        limit,
        id_contains: query,
        name_contains: query,
      })
      : await api.applications.list({ page, limit })
    return { applications: data.applications, totalCount: data.totalCount }
  },
})

const getApplicationsRightsLogic = createRequestLogic({
  type: applications.GET_APPS_RIGHTS_LIST,
  async process ({ action }) {
    const { id } = action.payload
    const result = await api.rights.applications(id)
    return result.rights.sort()
  },
})

export default [
  getApplicationsLogic,
  getApplicationsRightsLogic,
]
