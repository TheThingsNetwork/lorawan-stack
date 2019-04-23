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

import { createLogic } from 'redux-logic'

import api from '../../api'
import * as applications from '../actions/applications'
import * as application from '../actions/application'

const getApplicationsLogic = createLogic({
  type: [
    applications.GET_APPS_LIST,
    applications.SEARCH_APPS_LIST,
  ],
  latest: true,
  async process ({ getState, action }, dispatch, done) {
    const { page, pageSize: limit, query } = action.filters
    try {
      const data = query
        ? await api.applications.search({
          page,
          limit,
          id_contains: query,
          name_contains: query,
        })
        : await api.applications.list({ page, limit })
      dispatch(applications.getApplicationsSuccess(data.applications, data.totalCount))
    } catch (error) {
      dispatch(applications.getApplicationsFailure(error))
    }

    done()
  },
})

const getApplicationsRightsLogic = createLogic({
  type: [
    applications.GET_APPS_RIGHTS_LIST,
    application.GET_APP_API_KEY_PAGE_DATA,
    application.GET_APP_COLLABORATOR_PAGE_DATA,
  ],
  async process ({ getState, action }, dispatch, done) {
    const { id } = action
    try {
      const result = await api.rights.applications(id)

      dispatch(applications.getApplicationsRightsListSuccess(result.rights.sort()))
    } catch (error) {
      dispatch(applications.getApplicationsRightsListFailure(error))
    }

    done()
  },
})

export default [
  getApplicationsLogic,
  getApplicationsRightsLogic,
]
