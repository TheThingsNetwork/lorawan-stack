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

import tts from '@console/api/tts'

import { isNotFoundError } from '@ttn-lw/lib/errors/utils'
import createRequestLogic from '@ttn-lw/lib/store/logics/create-request-logic'

import * as repository from '@console/store/actions/device-repository'

const listDeviceBrandsLogic = createRequestLogic({
  type: repository.LIST_BRANDS,
  process: async ({ action }) => {
    const {
      payload: { appId, params = {} },
      meta: { selector = [] },
    } = action

    return tts.Applications.Devices.Repository.listBrands(appId, params, selector)
  },
})

const getDeviceBrandLogic = createRequestLogic({
  type: repository.GET_BRAND,
  process: ({ action }) => {
    const {
      payload: { appId, brandId },
      meta: { selector = [] },
    } = action

    return tts.Applications.Devices.Repository.getBrand(appId, brandId, selector)
  },
})

const listDeviceModelsLogic = createRequestLogic({
  type: repository.LIST_MODELS,
  process: async ({ action }) => {
    const {
      payload: { appId, brandId, params = {} },
      meta: { selector = [] },
    } = action

    const { models, totalCount } = await tts.Applications.Devices.Repository.listModels(
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

    return tts.Applications.Devices.Repository.getModel(appId, brandId, modelId, selector)
  },
})

const getTemplateLogic = createRequestLogic({
  type: repository.GET_TEMPLATE,
  process: ({ action }) => {
    const {
      payload: { appId, version },
    } = action

    return tts.Applications.Devices.Repository.getTemplate(appId, version)
  },
})

const getRepositoryPayloadFormattersLogic = createRequestLogic({
  type: repository.GET_REPO_PF,
  process: async ({ action }) => {
    const {
      payload: { appId, version },
    } = action

    let repositoryPayloadFormatters
    try {
      const uplinkDecoder = await tts.Applications.Devices.Repository.getUplinkDecoder(
        appId,
        version,
      )
      const downlinkDecoder = await tts.Applications.Devices.Repository.getDownlinkDecoder(
        appId,
        version,
      )
      const downlinkEncoder = await tts.Applications.Devices.Repository.getDownlinkEncoder(
        appId,
        version,
      )
      repositoryPayloadFormatters = {
        ...uplinkDecoder,
        ...downlinkDecoder,
        ...downlinkEncoder,
      }
      return repositoryPayloadFormatters
    } catch (error) {
      if (isNotFoundError(error) && typeof repositoryPayloadFormatters !== 'undefined') {
        return repositoryPayloadFormatters
      }

      throw error
    }
  },
})

export default [
  listDeviceBrandsLogic,
  getDeviceBrandLogic,
  listDeviceModelsLogic,
  getDeviceModelLogic,
  getTemplateLogic,
  getRepositoryPayloadFormattersLogic,
]
