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
import { push } from 'connected-react-router'
import Query from 'query-string'

import * as applications from '../actions/applications'

const PAGE_SIZE = 3
const apps = [
  {
    application_id: 'test-app',
    created_at: '2018-09-19T08:29:39.952Z',
    updated_at: '2018-09-19T08:29:39.952Z',
    name: 'Test App',
    description: 'description',
  },
  {
    application_id: 'test-app2',
    created_at: '2018-09-19T08:29:39.952Z',
    updated_at: '2018-09-19T08:29:39.952Z',
    name: 'Test App 2',
    description: 'description 2',
  },
  {
    application_id: 'test-app3',
    created_at: '2018-09-19T08:29:39.952Z',
    updated_at: '2018-09-19T08:29:39.952Z',
    name: 'Test App 3',
    description: 'description 3',
  },
  {
    application_id: 'test-app4',
    created_at: '2018-09-19T08:29:39.952Z',
    updated_at: '2018-09-19T08:29:39.952Z',
    name: 'Test App 4',
    description: 'description 4',
  },
  {
    application_id: 'test-app5',
    created_at: '2018-09-19T08:29:39.952Z',
    updated_at: '2018-09-19T08:29:39.952Z',
    name: 'Test App 5',
    description: 'description 5',
  },
  {
    application_id: 'test-app6',
    created_at: '2018-09-19T08:29:39.952Z',
    updated_at: '2018-09-19T08:29:39.952Z',
    name: 'Test App 6',
    description: 'description 6',
  },
]

const searchApplicationsStub = function (params) {
  const start = (params.page - 1) * PAGE_SIZE
  const end = start + PAGE_SIZE
  const query = params.query || ''

  const res = apps.filter(app => app.application_id.includes(query))
  const total = res.length

  return new Promise(resolve => setTimeout(() => resolve(
    { applications: res.slice(start, end), totalCount: total }
  ), 1000))
}

const getApplicationsStub = function (params) {
  const start = (params.page - 1) * PAGE_SIZE
  const end = start + PAGE_SIZE

  const res = apps.slice(start, end)
  const total = apps.length

  return new Promise(resolve => setTimeout(() => resolve(
    { applications: res, totalCount: total }
  ), 1000))
}

const DEFAULT_PAGE = 1
const DEFAULT_TAB = 'all'
const ALLOWED_TABS = [ 'all' ]
const ALLOWED_ORDERS = [ 'asc', 'desc', undefined ]

const transformParams = function ({ getState, action }, next) {
  const { type, ...rest } = action

  rest.query = Boolean(rest.query) ? rest.query : undefined

  if (!ALLOWED_TABS.includes(rest.tab)) {
    rest.tab = DEFAULT_TAB
  }

  if (!ALLOWED_ORDERS.includes(rest.order)) {
    rest.order = undefined
    rest.orderBy = undefined
  }

  if (
    Boolean(rest.order) && !Boolean(rest.orderBy)
      || !Boolean(rest.order) && Boolean(rest.orderBy)
  ) {
    rest.order = undefined
    rest.orderBy = undefined
  }

  if (!Boolean(rest.page) || rest.page < 0) {
    rest.page = DEFAULT_PAGE
  }

  next({ type, ...rest })
}

const changeSearchQueryLogic = createLogic({
  type: applications.CHANGE_APPS_SEARCH,
  async process ({ getState, action }, dispatch, done) {
    const { page, order, orderBy, tab } = getState().applications.filters

    dispatch(applications.searchApplicationsList(
      page,
      order,
      orderBy,
      tab,
      action.query
    ))

    done()
  },
})

const searchApplicationsLogic = createLogic({
  type: applications.SEARCH_APPS_LIST,
  debounce: 350,
  latest: true,
  transform: transformParams,
  async process ({ getState, action }, dispatch, done) {
    const { type, ...rest } = action
    const search = Query.stringify(rest)

    dispatch(push(`/console/applications?${search}`))

    try {
      const data = await searchApplicationsStub(rest)
      dispatch(applications.getApplicationsSuccess(data.applications, data.totalCount))
    } catch (error) {
      dispatch(applications.getApplicationsFailure(error))
    }

    done()
  },
})

const getApplicationsLogic = createLogic({
  type: [
    applications.GET_APPS_LIST,
    applications.CHANGE_APPS_ORDER,
    applications.CHANGE_APPS_PAGE,
    applications.CHANGE_APPS_TAB,
  ],
  transform: transformParams,
  async process ({ getState, action }, dispatch, done) {
    const { type, ...rest } = action
    const search = Query.stringify(rest)

    dispatch(push(`/console/applications?${search}`))

    try {
      const data = await getApplicationsStub(rest)
      dispatch(applications.getApplicationsSuccess(data.applications, data.totalCount))
    } catch (error) {
      dispatch(applications.getApplicationsFailure(error))
    }

    done()
  },
})

export default [
  getApplicationsLogic,
  searchApplicationsLogic,
  changeSearchQueryLogic,
]
