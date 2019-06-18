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

import { GET_APPS_LIST_BASE } from '../actions/applications'
import { GET_APP_BASE } from '../actions/application'
import {
  createPaginationIdsSelectorByEntity,
  createPaginationTotalCountSelectorByEntity,
} from './pagination'
import { createFetchingSelector } from './fetching'
import { createErrorSelector } from './error'

const ENTITY = 'applications'

// Application
export const selectApplicationStore = state => state.applications
export const selectApplicationEntitiesStore = state => selectApplicationStore(state).entities
export const selectApplicationById = (state, id) => selectApplicationEntitiesStore(state)[id]
export const selectSelectedApplicationId = state => selectApplicationStore(state).selectedApplication
export const selectSelectedApplication = state => selectApplicationById(state, selectSelectedApplicationId(state))
export const selectApplicationFetching = createFetchingSelector(GET_APP_BASE)
export const selectApplicationError = createErrorSelector(GET_APP_BASE)

// Applications
const selectAppsIds = createPaginationIdsSelectorByEntity(ENTITY)
const selectAppsTotalCount = createPaginationTotalCountSelectorByEntity(ENTITY)
const selectAppsFetching = createFetchingSelector(GET_APPS_LIST_BASE)
const selectAppsError = createErrorSelector(GET_APPS_LIST_BASE)

export const selectApplications = state => selectAppsIds(state).map(id => selectApplicationById(state, id))
export const selectApplicationsTotalCount = state => selectAppsTotalCount(state)
export const selectApplicationsFetching = state => selectAppsFetching(state)
export const selectApplicationsError = state => selectAppsError(state)
