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

import createRequestActions from '@ttn-lw/lib/store/actions/create-request-actions'

export const createPaginationBaseActionType = name => `GET_${name}_LIST`

export const createPaginationDeleteBaseActionType = name => `DELETE_${name}`

export const createPaginationRestoreBaseActionType = name => `RESTORE_${name}`

export const createPaginationByParentRequestActions = name =>
  createRequestActions(
    createPaginationBaseActionType(name),
    (parentType, parentId, { page, limit, query, order } = {}) => ({
      parentType,
      parentId,
      params: { page, limit, query, order },
    }),
    (parentType, parentId, params, selectors = []) => ({ selectors }),
  )

export const createPaginationByIdRequestActions = (
  name,
  requestPayloadCreator = (id, { page, limit, query, order } = {}) => ({
    id,
    params: { page, limit, query, order },
  }),
  requestMetaCreator = (id, params, selectors = [], options = {}) => ({ selectors, options }),
) =>
  createRequestActions(
    createPaginationBaseActionType(name),
    requestPayloadCreator,
    requestMetaCreator,
  )

export const createPaginationRequestActions = (
  name,
  requestPayloadCreator = ({ page, limit, query, order, deleted } = {}) => ({
    params: { page, limit, query, order, deleted },
  }),
  requestMetaCreator = (params, selectors = [], options = {}) => ({ selectors, options }),
) =>
  createRequestActions(
    createPaginationBaseActionType(name),
    requestPayloadCreator,
    requestMetaCreator,
  )

export const createPaginationDeleteActions = name =>
  createRequestActions(
    createPaginationDeleteBaseActionType(name),
    id => ({ id }),
    (id, options = {}) => ({ options }),
  )

export const createPaginationByIdDeleteActions = name =>
  createRequestActions(createPaginationDeleteBaseActionType(name), (id, targetId) => ({
    id,
    targetId,
  }))

export const createPaginationByRouteParametersDeleteActions = name =>
  createRequestActions(createPaginationDeleteBaseActionType(name), (routeParams, id) => ({
    routeParams,
    id,
  }))

export const createPaginationRestoreActions = name =>
  createRequestActions(
    createPaginationRestoreBaseActionType(name),
    id => ({ id }),
    (id, options = {}) => ({ options }),
  )
