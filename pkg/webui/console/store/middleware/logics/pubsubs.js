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

import * as pubsubs from '../../actions/pubsubs'

import api from '../../../api'
import createRequestLogic from './lib'

const getPubsubLogic = createRequestLogic({
  type: pubsubs.GET_PUBSUB,
  async process({ action }) {
    const {
      payload: { appId, pubsubId },
      meta: { selector },
    } = action
    return api.application.pubsubs.get(appId, pubsubId, selector)
  },
})

const getPubsubsLogic = createRequestLogic({
  type: pubsubs.GET_PUBSUBS_LIST,
  async process({ action }) {
    const { appId } = action.payload
    const res = await api.application.pubsubs.list(appId)
    return { entities: res.pubsubs, totalCount: res.totalCount }
  },
})

export default [getPubsubLogic, getPubsubsLogic]
