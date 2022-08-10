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

import tts from '@console/api/tts'

import { isNotFoundError } from '@ttn-lw/lib/errors/utils'
import createRequestLogic from '@ttn-lw/lib/store/logics/create-request-logic'

import {
  GET_APP_PKG_DEFAULT_ASSOC,
  SET_APP_PKG_DEFAULT_ASSOC,
  DELETE_APP_PKG_DEFAULT_ASSOC,
} from '@console/store/actions/application-packages'

const getApplicationPackagesDefaultAssociationLogic = createRequestLogic({
  type: GET_APP_PKG_DEFAULT_ASSOC,
  process: async ({ action }) => {
    const { appId, fPort } = action.payload
    const { selector } = action.meta
    try {
      const result = await tts.Applications.Packages.getDefaultAssociation(appId, fPort, selector)

      return result
    } catch (error) {
      if (isNotFoundError(error)) {
        // 404s are expected when the default package does not exist. This should not
        // result in a failure action.
        return { ids: { f_port: fPort } }
      }
      throw error
    }
  },
})

const setApplicationPackagesDefaultAssociationLogic = createRequestLogic({
  type: SET_APP_PKG_DEFAULT_ASSOC,
  process: ({ action }) => {
    const { appId, fPort, data } = action.payload
    const { selector } = action.meta

    return tts.Applications.Packages.setDefaultAssociation(appId, fPort, data, selector)
  },
})

const deleteApplicationPackagesDefaultAssociationLogic = createRequestLogic({
  type: DELETE_APP_PKG_DEFAULT_ASSOC,
  process: async ({ action }) => {
    const { appId, fPort } = action.payload
    await tts.Applications.Packages.deleteDefaultAssociation(appId, fPort)

    return { fPort }
  },
})

export default [
  getApplicationPackagesDefaultAssociationLogic,
  setApplicationPackagesDefaultAssociationLogic,
  deleteApplicationPackagesDefaultAssociationLogic,
]
