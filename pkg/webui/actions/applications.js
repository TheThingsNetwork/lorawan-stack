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

export const GET_APPS_LIST = 'GET_APPLICATIONS_LIST'
export const SEARCH_APPS_LIST = 'SEARCH_APPLICATIONS_LIST'
export const GET_APPS_LIST_SUCCESS = 'GET_APPLICATIONS_LIST_SUCCESS'
export const GET_APPS_LIST_FAILURE = 'GET_APPLICATIONS_LIST_FAILURE'
export const CHANGE_APPS_PAGE = 'CHANGE_APPLICATIONS_PAGE'
export const CHANGE_APPS_ORDER = 'CHANGE_APPLICATIONS_ORDER'
export const CHANGE_APPS_TAB = 'CHANGE_APPLICATIONS_TAB'
export const CHANGE_APPS_SEARCH = 'CHANGE_APPLICATIONS_SEARCH'

export const getApplicationsList = filters => (
  { type: GET_APPS_LIST, filters }
)

export const searchApplicationsList = filters => (
  { type: SEARCH_APPS_LIST, filters }
)

export const getApplicationsSuccess = (applications, totalCount) => (
  { type: GET_APPS_LIST_SUCCESS, applications, totalCount }
)

export const getApplicationsFailure = error => (
  { type: GET_APPS_LIST_FAILURE, error }
)

export const changeApplicationsPage = page => (
  { type: CHANGE_APPS_PAGE, page }
)

export const changeApplicationsOrder = (order, orderBy) => (
  { type: CHANGE_APPS_ORDER, order, orderBy }
)

export const changeApplicationsTab = tab => (
  { type: CHANGE_APPS_TAB, tab }
)

export const changeApplicationsSearch = query => (
  { type: CHANGE_APPS_SEARCH, query }
)
