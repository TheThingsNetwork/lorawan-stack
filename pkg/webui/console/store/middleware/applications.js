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
import { replace } from 'connected-react-router'

import api from '../../api'
import * as applications from '../actions/applications'

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

const deleteApplicationLogic = createLogic({
  type: applications.DELETE_APP,
  async process ({ getState, action }, dispatch, done) {
    const id = action.id

    try {
      await api.application.delete(id)
      dispatch(applications.deleteApplicationSuccess())
      dispatch(replace('/console/applications'))
    } catch (error) {
      dispatch(applications.deleteApplicationFailure(error))
    }

    done()
  },
})

const updateApplicationLogic = createLogic({
  type: applications.UPDATE_APP,
  async process ({ getState, action }, dispatch, done) {
    const data = action.data
    const application = getState().application.application

    const changed = Object.keys(data).reduce(function (patch, field) {
      const oldValue = application[field]
      const newValue = data[field]

      if (oldValue !== newValue) {
        patch[field] = newValue
      }

      return patch
    }, {})

    try {
      await api.application.update(application.ids.application_id, changed)
      dispatch(applications.updateApplicationSuccess())
    } catch (error) {
      dispatch(applications.updateApplicationFailure(error))
    }

    done()
  },
})

const createApplicationLogic = createLogic({
  type: applications.CREATE_APP,
  async process ({ getState, action }, dispatch, done) {
    const application = action.data
    const userId = getState().user.user.ids.user_id

    try {
      const result = await api.application.create(userId,
        {
          ids: { application_id: application.application_id },
          name: application.name,
          description: application.description,
        })
      dispatch(applications.createApplicationSuccess(result))
      dispatch(replace('/console/applications'))
    } catch (error) {
      dispatch(applications.createApplicationFailure(error))
    }

    done()
  },
})

const getApplicationsRightsLogic = createLogic({
  type: applications.GET_APPS_RIGHTS_LIST,
  async process ({ getState, action }, dispatch, done) {
    try {
      const rights = await api.rights.applications()

      dispatch(applications.getApplicationsRightsListSuccess(rights))
    } catch (error) {
      dispatch(applications.getApplicationsRightsListFailure(error))
    }

    done()
  },
})

export default [
  getApplicationsLogic,
  deleteApplicationLogic,
  createApplicationLogic,
  updateApplicationLogic,
  getApplicationsRightsLogic,
]
