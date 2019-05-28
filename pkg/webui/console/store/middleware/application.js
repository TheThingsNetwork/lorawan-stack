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

import { isNotFoundError } from '../../../lib/errors/utils'
import api from '../../api'
import * as application from '../actions/application'
import * as link from '../actions/link'
import createEventsConnectLogics from './events'

const getApplicationLogic = createLogic({
  type: [ application.GET_APP ],
  async process ({ getState, action }, dispatch, done) {
    const { id } = action
    try {
      const app = await api.application.get(id, 'name,description')
      dispatch(application.getApplicationSuccess(app))
    } catch (e) {
      dispatch(application.getApplicationFailure(e))
    }

    done()
  },
})

const getApplicationApiKeysLogic = createLogic({
  type: application.GET_APP_API_KEYS_LIST,
  async process ({ getState, action }, dispatch, done) {
    const { id, params } = action
    try {
      const res = await api.application.apiKeys.list(id, params)
      dispatch(
        application.getApplicationApiKeysListSuccess(
          id,
          res.api_keys,
          res.totalCount
        )
      )
    } catch (e) {
      dispatch(application.getApplicationApiKeysListFailure(id, e))
    }

    done()
  },
})

const getApplicationApiKeyLogic = createLogic({
  type: application.GET_APP_API_KEY,
  async process ({ action }, dispatch, done) {
    const { entityId, keyId } = action
    try {
      const key = await api.application.apiKeys.get(entityId, keyId)
      dispatch(application.getApplicationApiKeySuccess(key))
    } catch (error) {
      dispatch(application.getApplicationApiKeyFailure(error))
    }

    done()
  },
})

const getApplicationCollaboratorsLogic = createLogic({
  type: [
    application.GET_APP_COLLABORATOR_PAGE_DATA,
    application.GET_APP_COLLABORATORS_LIST,
  ],
  async process ({ getState, action }, dispatch, done) {
    const { id } = action
    try {
      const res = await api.application.collaborators.list(id)
      const collaborators = res.collaborators.map(function (collaborator) {
        const { ids, ...rest } = collaborator
        const isUser = !!ids.user_ids
        const collaboratorId = isUser
          ? ids.user_ids.user_id
          : ids.organization_ids.organization_id

        return {
          id: collaboratorId,
          isUser,
          ...rest,
        }
      })

      dispatch(
        application.getApplicationCollaboratorsListSuccess(
          id,
          collaborators,
          res.totalCount
        )
      )
    } catch (e) {
      dispatch(application.getApplicationCollaboratorsListFailure(id, e))
    }

    done()
  },
})

const getApplicationLinkLogic = createLogic({
  type: link.GET_APP_LINK,
  async process ({ action }, dispatch, done) {
    const { id, meta = {}} = action
    const { selectors = []} = meta

    let linkResult
    let statsResult
    try {
      linkResult = await api.application.link.get(id, selectors)
      statsResult = await api.application.link.stats(id)

      dispatch(link.getApplicationLinkSuccess(linkResult, statsResult, true))
    } catch (error) {
      // consider errors that are not 404, since not found means that the
      // application is not linked.
      if (isNotFoundError(error)) {
        dispatch(link.getApplicationLinkSuccess(linkResult, statsResult, false))
      } else {
        dispatch(link.getApplicationLinkFailure(error))
      }
    }

    done()
  },
})

export default [
  getApplicationLogic,
  getApplicationApiKeysLogic,
  getApplicationApiKeyLogic,
  getApplicationCollaboratorsLogic,
  ...createEventsConnectLogics(application.SHARED_NAME, 'application'),
  getApplicationLinkLogic,
]
