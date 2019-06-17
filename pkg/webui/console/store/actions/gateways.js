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

import createGetRightsListRequestActions, { createGetRightsListActionType } from './rights'
import { createRequestActions } from './lib'

export const SHARED_NAME = 'GATEWAYS'

export const GET_GTWS_LIST_BASE = 'GET_GATEWAYS_LIST'
export const [{
  request: GET_GTWS_LIST,
  success: GET_GTWS_LIST_SUCCESS,
  failure: GET_GTWS_LIST_FAILURE,
}, {
  request: getGatewaysList,
  success: getGatewaysSuccess,
  failure: getGatetaysFailure,
}] = createRequestActions(GET_GTWS_LIST_BASE,
  ({ page, limit, query } = {}) => ({ page, limit, query })
)

export const GET_GTWS_RIGHTS_LIST_BASE = createGetRightsListActionType(SHARED_NAME)
export const [{
  request: GET_GTWS_RIGHTS_LIST,
  success: GET_GTWS_RIGHTS_LIST_SUCCESS,
  failure: GET_GTWS_RIGHTS_LIST_FAILURE,
}, {
  request: getGatewaysRightsList,
  success: getGatewaysRightsListSuccess,
  failure: getGatewaysRightsListFailure,
}] = createGetRightsListRequestActions(SHARED_NAME)
