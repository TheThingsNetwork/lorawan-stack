// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

import createRequestActions from '@ttn-lw/lib/store/actions/create-request-actions'

export const GET_APP_PKG_DEFAULT_ASSOC_BASE = 'GET_APPLICATION_PACKAGE_DEFAULT_ASSOCIATION'
export const [
  {
    request: GET_APP_PKG_DEFAULT_ASSOC,
    success: GET_APP_PKG_DEFAULT_ASSOC_SUCCESS,
    failure: GET_APP_PKG_DEFAULT_ASSOC_FAILURE,
  },
  {
    request: getAppPkgDefaultAssoc,
    success: getAppPkgDefaultAssocSuccess,
    failure: getAppPkgDefaultAssocFailure,
  },
] = createRequestActions(
  GET_APP_PKG_DEFAULT_ASSOC_BASE,
  (appId, fPort) => ({ appId, fPort }),
  (appId, fPort, selector) => ({ selector }),
)

export const SET_APP_PKG_DEFAULT_ASSOC_BASE = 'SET_APPLICATION_PACKAGE_DEFAULT_ASSOCIATION'
export const [
  {
    request: SET_APP_PKG_DEFAULT_ASSOC,
    success: SET_APP_PKG_DEFAULT_ASSOC_SUCCESS,
    failure: SET_APP_PKG_DEFAULT_ASSOC_FAILURE,
  },
  {
    request: setAppPkgDefaultAssoc,
    success: setAppPkgDefaultAssocSuccess,
    failure: setAppPkgDefaultAssocFailure,
  },
] = createRequestActions(
  SET_APP_PKG_DEFAULT_ASSOC_BASE,
  (appId, fPort, data) => ({ appId, fPort, data }),
  (appId, fPort, data, selector) => ({ selector }),
)
