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

import * as configuration from '../actions/configuration'
import { get as cacheGet, set as cacheSet } from '../../lib/cache'
import api from '../../api'

import {
  nsFrequencyPlansSelector,
  gsFrequencyPlansSelector,
} from '../selectors/configuration'

const getNsFrequencyPlansLogic = createLogic({
  type: configuration.GET_NS_FREQUENCY_PLANS,
  validate ({ getState, action }, allow, reject) {
    const plansNs = nsFrequencyPlansSelector(getState())
    if (plansNs && plansNs.length) {
      reject()
    } else {
      allow(action)
    }
  },
  async process ({ getState, action }, dispatch, done) {
    let frequencyPlans = cacheGet('ns_frequency_plans')
    try {
      if (!frequencyPlans) {
        frequencyPlans = (await api.configuration.listNsFrequencyPlans()).frequency_plans
        cacheSet('ns_frequency_plans', frequencyPlans)
      }
      dispatch(configuration.getNsFrequencyPlansSuccess(frequencyPlans))
    } catch (error) {
      dispatch(configuration.getNsFrequencyPlansFailure(error))
    }

    done()
  },
})

const getGsFrequencyPlansLogic = createLogic({
  type: configuration.GET_GS_FREQUENCY_PLANS,
  validate ({ getState, action }, allow, reject) {
    const plansGs = gsFrequencyPlansSelector(getState())
    if (plansGs && plansGs.length) {
      reject()
    } else {
      allow(action)
    }
  },
  async process ({ getState, action }, dispatch, done) {
    let frequencyPlans = cacheGet('gs_frequency_plans')
    try {
      if (!frequencyPlans) {
        frequencyPlans = (await api.configuration.listGsFrequencyPlans()).frequency_plans
        cacheSet('gs_frequency_plans', frequencyPlans)
      }
      dispatch(configuration.getGsFrequencyPlansSuccess(frequencyPlans))
    } catch (error) {
      dispatch(configuration.getGsFrequencyPlansFailure(error))
    }

    done()
  },
})

export default [
  getNsFrequencyPlansLogic,
  getGsFrequencyPlansLogic,
]
