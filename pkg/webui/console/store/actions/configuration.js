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

export const GET_NS_FREQUENCY_PLANS = 'GET_NS_FREQUENCY_PLANS'
export const GET_NS_FREQUENCY_PLANS_SUCCESS = 'GET_NS_FREQUENCY_PLANS_SUCCESS'
export const GET_NS_FREQUENCY_PLANS_FAILURE = 'GET_NS_FREQUENCY_PLANS_FAILURE'
export const GET_GS_FREQUENCY_PLANS = 'GET_GS_FREQUENCY_PLANS'
export const GET_GS_FREQUENCY_PLANS_SUCCESS = 'GET_GS_FREQUENCY_PLANS_SUCCESS'
export const GET_GS_FREQUENCY_PLANS_FAILURE = 'GET_GS_FREQUENCY_PLANS_FAILURE'

export const getNsFrequencyPlans = () => (
  { type: GET_NS_FREQUENCY_PLANS }
)

export const getNsFrequencyPlansSuccess = frequencyPlans => (
  { type: GET_NS_FREQUENCY_PLANS_SUCCESS, frequencyPlans }
)

export const getNsFrequencyPlansFailure = error => (
  { type: GET_NS_FREQUENCY_PLANS_FAILURE, error }
)

export const getGsFrequencyPlans = () => (
  { type: GET_GS_FREQUENCY_PLANS }
)

export const getGsFrequencyPlansSuccess = frequencyPlans => (
  { type: GET_GS_FREQUENCY_PLANS_SUCCESS, frequencyPlans }
)

export const getGsFrequencyPlansFailure = error => (
  { type: GET_GS_FREQUENCY_PLANS_FAILURE, error }
)
