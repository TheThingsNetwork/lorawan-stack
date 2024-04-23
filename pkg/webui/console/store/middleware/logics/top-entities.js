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

import createRequestLogic from '@ttn-lw/lib/store/logics/create-request-logic'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import { getTopFrequencyRecencyItems } from '@console/lib/frequently-visited-entities'

import * as actions from '@console/store/actions/top-entities'
import { getBookmarksList } from '@console/store/actions/user-preferences'

const getTopEntitiesLogic = createRequestLogic({
  type: actions.GET_TOP_ENTITIES,
  process: async ({ action }, dispatch) => {
    const topFrequencyRecencyItems = getTopFrequencyRecencyItems()

    return topFrequencyRecencyItems
  },
})

export default [getTopEntitiesLogic]
