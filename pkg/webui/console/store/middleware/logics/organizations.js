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

import tts from '@console/api/tts'

import createRequestLogic from '@ttn-lw/lib/store/logics/create-request-logic'
import { getOrganizationId } from '@ttn-lw/lib/selectors/id'

import * as organizations from '@console/store/actions/organizations'

import { selectUserId } from '@console/store/selectors/user'

import createEventsConnectLogics from './events'

const getOrganizationLogic = createRequestLogic({
  type: organizations.GET_ORG,
  process: async ({ action }, dispatch) => {
    const {
      payload: { id },
      meta: { selector },
    } = action
    const org = await tts.Organizations.getById(id, selector)
    dispatch(organizations.startOrganizationEventsStream(id))
    return org
  },
})

const getOrganizationsLogic = createRequestLogic({
  type: organizations.GET_ORGS_LIST,
  latest: true,
  process: async ({ action }, dispatch) => {
    const {
      params: { page, limit, order, query, deleted },
    } = action.payload
    const { selectors, options } = action.meta

    const data = options.isSearch
      ? await tts.Organizations.search(
          {
            page,
            limit,
            query,
            deleted,
          },
          selectors,
        )
      : await tts.Organizations.getAll({ page, limit, order }, selectors)

    if (options.withCollaboratorCount) {
      for (const org of data.organizations) {
        dispatch(organizations.getOrganizationCollaboratorCount(getOrganizationId(org)))
      }
    }

    return {
      entities: data.organizations,
      totalCount: data.totalCount,
    }
  },
})

const getOrganizationsCollaboratorCountLogic = createRequestLogic({
  type: organizations.GET_ORG_COLLABORATOR_COUNT,
  process: async ({ action }) => {
    const { id: orgId } = action.payload
    const result = await tts.Organizations.Collaborators.getAll(orgId, { limit: 1 })

    return { id: orgId, collaboratorCount: result.totalCount }
  },
})

const createOrganizationLogic = createRequestLogic({
  type: organizations.CREATE_ORG,
  process: async ({ action, getState }) => {
    const userId = selectUserId(getState())

    return tts.Organizations.create(userId, action.payload)
  },
})

const updateOrganizationLogic = createRequestLogic({
  type: organizations.UPDATE_ORG,
  process: async ({ action }) => {
    const { id, patch } = action.payload

    const result = await tts.Organizations.updateById(id, patch)

    return { ...patch, ...result }
  },
})

const deleteOrganizationLogic = createRequestLogic({
  type: organizations.DELETE_ORG,
  process: async ({ action }) => {
    const { id } = action.payload
    const { options } = action.meta

    if (options.purge) {
      await tts.Organizations.purgeById(id)
    } else {
      await tts.Organizations.deleteById(id)
    }

    return { id }
  },
})

const restoreOrganizationLogic = createRequestLogic({
  type: organizations.RESTORE_ORG,
  process: async ({ action }) => {
    const { id } = action.payload

    await tts.Organizations.restoreById(id)

    return { id }
  },
})

const getOrganizationsRightsLogic = createRequestLogic({
  type: organizations.GET_ORGS_RIGHTS_LIST,
  process: async ({ action }) => {
    const { id } = action.payload
    const result = await tts.Organizations.getRightsById(id)
    return result.rights.sort()
  },
})

export default [
  getOrganizationLogic,
  getOrganizationsLogic,
  getOrganizationsCollaboratorCountLogic,
  createOrganizationLogic,
  updateOrganizationLogic,
  deleteOrganizationLogic,
  restoreOrganizationLogic,
  getOrganizationsRightsLogic,
  ...createEventsConnectLogics(
    organizations.SHARED_NAME,
    'organizations',
    tts.Organizations.openStream,
  ),
]
