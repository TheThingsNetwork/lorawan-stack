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

import { createRequestActions } from './lib'

export const createPaginationBaseActionType = name => `GET_${name}_LIST`

export const createPaginationDeleteBaseActionType = name => `DELETE_${name}`

export const createPaginationByIdRequestActions = name =>
  createRequestActions(
    createPaginationBaseActionType(name),
    (id, { page, limit, query } = {}) => ({ id, params: { page, limit, query } }),
    (id, params, selectors = []) => ({ selectors }),
  )

export const createPaginationRequestActions = name =>
  createRequestActions(
    createPaginationBaseActionType(name),
    ({ page, limit, query } = {}) => ({ params: { page, limit, query } }),
    (params, selectors) => ({ selectors }),
  )

export const createPaginationDeleteActions = name =>
  createRequestActions(createPaginationDeleteBaseActionType(name), id => ({ id }))

export const createPaginationByIdDeleteActions = name =>
  createRequestActions(createPaginationDeleteBaseActionType(name), (id, targetId) => ({
    id,
    targetId,
  }))
