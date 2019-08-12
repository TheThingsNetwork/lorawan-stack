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

import * as configuration from '../../actions/configuration'
import api from '../../../api'

import { selectNsFrequencyPlans, selectGsFrequencyPlans } from '../../selectors/configuration'

import createRequestLogic from './lib'

const getNsFrequencyPlansLogic = createRequestLogic({
  type: configuration.GET_NS_FREQUENCY_PLANS,
  validate({ getState, action }, allow, reject) {
    const plansNs = selectNsFrequencyPlans(getState())
    if (plansNs && plansNs.length) {
      reject()
    } else {
      allow(action)
    }
  },
  async process() {
    const frequencyPlans = (await api.configuration.listNsFrequencyPlans()).frequency_plans

    return frequencyPlans
  },
})

const getGsFrequencyPlansLogic = createRequestLogic({
  type: configuration.GET_GS_FREQUENCY_PLANS,
  validate({ getState, action }, allow, reject) {
    const plansGs = selectGsFrequencyPlans(getState())
    if (plansGs && plansGs.length) {
      reject()
    } else {
      allow(action)
    }
  },
  async process() {
    const frequencyPlans = (await api.configuration.listGsFrequencyPlans()).frequency_plans

    return frequencyPlans
  },
})

export default [getNsFrequencyPlansLogic, getGsFrequencyPlansLogic]
