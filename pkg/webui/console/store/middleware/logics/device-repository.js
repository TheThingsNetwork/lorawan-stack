// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

import api from '@console/api'

import createRequestLogic from '@ttn-lw/lib/store/logics/create-request-logic'

import * as repository from '@console/store/actions/device-repository'

const listDeviceBrandsLogic = createRequestLogic({
  type: repository.LIST_BRANDS,
  process: async ({ action }) => {
    const {
      payload: { appId, params = {} },
      meta: { selector = [] },
    } = action

    return api.deviceRepository.listBrands(appId, params, selector)
  },
})

const getDeviceBrandLogic = createRequestLogic({
  type: repository.GET_BRAND,
  process: ({ action }) => {
    const {
      payload: { appId, brandId },
      meta: { selector = [] },
    } = action

    return api.deviceRepository.getBrand(appId, brandId, selector)
  },
})

const listDeviceModelsLogic = createRequestLogic({
  type: repository.LIST_MODELS,
  process: async ({ action }) => {
    const {
      payload: { appId, brandId, params = {} },
      meta: { selector = [] },
    } = action

    const { models, totalCount } = await api.deviceRepository.listModels(
      appId,
      brandId,
      params,
      selector,
    )

    return { models, totalCount, brandId }
  },
})

const getDeviceModelLogic = createRequestLogic({
  type: repository.GET_MODEL,
  process: ({ action }) => {
    const {
      payload: { appId, brandId, modelId },
      meta: { selector = [] },
    } = action

    return api.deviceRepository.getModel(appId, brandId, modelId, selector)
  },
})

const getTemplateLogic = createRequestLogic({
  type: repository.GET_TEMPLATE,
  process: ({ action }) => {
    const {
      payload: { appId, version },
    } = action

    return api.deviceRepository.getTemplate(appId, version)
  },
})

export default [
  listDeviceBrandsLogic,
  getDeviceBrandLogic,
  listDeviceModelsLogic,
  getDeviceModelLogic,
  getTemplateLogic,
]
