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

import {
  getRightsList,
  createGetRightsListActionType,
  getRightsListFailure,
  createGetRightsListFailureActionType,
  getRightsListSuccess,
  createGetRightsListSuccessActionType,
} from '../actions/rights'

export const SHARED_NAME = 'APPLICATIONS'

export const GET_APPS_LIST = 'GET_APPLICATIONS_LIST'
export const SEARCH_APPS_LIST = 'SEARCH_APPLICATIONS_LIST'
export const GET_APPS_LIST_SUCCESS = 'GET_APPLICATIONS_LIST_SUCCESS'
export const GET_APPS_LIST_FAILURE = 'GET_APPLICATIONS_LIST_FAILURE'
export const GET_APPS_RIGHTS_LIST = createGetRightsListActionType(SHARED_NAME)
export const GET_APPS_RIGHTS_LIST_SUCCESS = createGetRightsListSuccessActionType(SHARED_NAME)
export const GET_APPS_RIGHTS_LIST_FAILURE = createGetRightsListFailureActionType(SHARED_NAME)

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

export const getApplicationsRightsList = getRightsList(SHARED_NAME)

export const getApplicationsRightsListSuccess = getRightsListSuccess(SHARED_NAME)

export const getApplicationsRightsListFailure = getRightsListFailure(SHARED_NAME)
