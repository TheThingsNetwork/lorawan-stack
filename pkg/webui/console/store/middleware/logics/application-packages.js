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

import api from '@console/api'

import {
  GET_APP_PKG_DEFAULT_ASSOC,
  SET_APP_PKG_DEFAULT_ASSOC,
} from '@console/store/actions/application-packages'

import createRequestLogic from './lib'

const getApplicationPackagesDefaultAssociationLogic = createRequestLogic({
  type: GET_APP_PKG_DEFAULT_ASSOC,
  process({ action }) {
    const { appId, fPort } = action.payload
    const { selector } = action.meta
    return api.application.packages.getDefaultAssociation(appId, fPort, selector)
  },
})

const setApplicationPackagesDefaultAssociationLogic = createRequestLogic({
  type: SET_APP_PKG_DEFAULT_ASSOC,
  process({ action }) {
    const { appId, fPort, data } = action.payload
    const { selector } = action.meta
    return api.application.packages.setDefaultAssociation(appId, fPort, data, selector)
  },
})

export default [
  getApplicationPackagesDefaultAssociationLogic,
  setApplicationPackagesDefaultAssociationLogic,
]
