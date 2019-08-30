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

import api from '../../../api'
import * as organizations from '../../actions/organizations'
import createRequestLogic from './lib'

const getOrganizationsLogic = createRequestLogic({
  type: organizations.GET_ORGS_LIST,
  latest: true,
  async process({ action }) {
    const {
      params: { page, limit },
    } = action.payload
    const { selectors } = action.meta

    const data = await api.organizations.list({ page, limit }, selectors)

    return {
      entities: data.organizations,
      totalCount: data.totalCount,
    }
  },
})

export default [getOrganizationsLogic]
