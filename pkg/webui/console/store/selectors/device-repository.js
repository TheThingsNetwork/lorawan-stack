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

import { createFetchingSelector } from '@ttn-lw/lib/store/selectors/fetching'
import { createErrorSelector } from '@ttn-lw/lib/store/selectors/error'

import {
  LIST_BRANDS_BASE,
  LIST_MODELS_BASE,
  GET_TEMPLATE_BASE,
} from '@console/store/actions/device-repository'

const selectDRStore = store => store.deviceRepository

// Brands.

export const selectDeviceBrands = state => selectDRStore(state).brands.list
export const selectDeviceBrandsFetching = createFetchingSelector(LIST_BRANDS_BASE)
export const selectDeviceBrandsError = createErrorSelector(LIST_BRANDS_BASE)

// Models.

export const selectDeviceModelsByBrandId = (state, brandId) => {
  const models = selectDRStore(state).models[brandId] || { list: [] }

  return models.list
}
export const selectDeviceModelsFetching = createFetchingSelector(LIST_MODELS_BASE)
export const selectDeviceModelsError = createErrorSelector(LIST_MODELS_BASE)
export const selectDeviceModelById = (state, brandId, modelId) => {
  const models = selectDeviceModelsByBrandId(state, brandId)

  return models.find(model => model.model_id === modelId)
}
export const selectDeviceModelHardwareVersions = (state, brandId, modelId) => {
  const model = selectDeviceModelById(state, brandId, modelId) || {}
  if (!model.hardware_versions) {
    return []
  }

  return model.hardware_versions
}
export const selectDeviceModelFirmwareVersions = (state, brandId, modelId) => {
  const model = selectDeviceModelById(state, brandId, modelId) || {}
  if (!model.firmware_versions) {
    return []
  }

  return model.firmware_versions
}

// Template.

export const selectDeviceTemplate = state => selectDRStore(state).template
export const selectDeviceTemplateFetching = createFetchingSelector(GET_TEMPLATE_BASE)
export const selectDeviceTemplateError = createErrorSelector(GET_TEMPLATE_BASE)

export const selectDeviceRepoPayloadFromatters = state =>
  selectDRStore(state).repo_payload_formatters
