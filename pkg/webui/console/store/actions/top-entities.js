// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

export const GET_TOP_ENTITIES_BASE = 'GET_TOP_ENTITIES'
export const [
  {
    request: GET_TOP_ENTITIES,
    success: GET_TOP_ENTITIES_SUCCESS,
    failure: GET_TOP_ENTITIES_FAILURE,
  },
  { request: getTopEntities, success: getTopEntitiesSuccess, failure: getTopEntitiesFailure },
] = createRequestActions(GET_TOP_ENTITIES_BASE, params => ({ ...params }))
