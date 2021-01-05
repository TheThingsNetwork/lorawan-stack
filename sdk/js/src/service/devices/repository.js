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

import Marshaler from '../../util/marshaler'

class DeviceRepository {
  constructor(registry) {
    this._api = registry
  }

  // Brands retrieval.

  async listBrands(appId, params = {}, selector = []) {
    const result = await this._api.ListBrands(
      {
        routeParams: {
          'application_ids.application_id': appId,
        },
      },
      { ...params, ...Marshaler.selectorToFieldMask(selector) },
    )

    return Marshaler.payloadListResponse('brands', result)
  }

  async getBrand(appId, brandId, selector = []) {
    const result = await this._api.GetBrand(
      {
        routeParams: {
          'application_ids.application_id': appId,
          brand_id: brandId,
        },
      },
      { ...Marshaler.selectorToFieldMask(selector) },
    )

    return Marshaler.payloadSingleResponse(result)
  }

  // Models retrieval.

  async listModels(appId, brandId, params = {}, selector = []) {
    const result = await this._api.ListModels(
      {
        routeParams: {
          'application_ids.application_id': appId,
          brand_id: brandId,
        },
      },
      { ...params, ...Marshaler.selectorToFieldMask(selector) },
    )

    return Marshaler.payloadListResponse('models', result)
  }

  async getModel(appId, brandId, modelId, selector = []) {
    const result = await this._api.GetModel(
      {
        routeParams: {
          'application_ids.application_id': appId,
          brand_id: brandId,
          model_id: modelId,
        },
      },
      {
        ...Marshaler.selectorToFieldMask(selector),
      },
    )

    return Marshaler.payloadSingleResponse(result)
  }

  // Templates retrieval.

  async getTemplate(appId, version) {
    const result = await this._api.GetTemplate({
      routeParams: {
        'application_ids.application_id': appId,
        'version_ids.brand_id': version.brand_id,
        'version_ids.model_id': version.model_id,
        'version_ids.firmware_version': version.firmware_version,
        'version_ids.band_id': version.band_id,
      },
    })

    return Marshaler.payloadSingleResponse(result)
  }

  // Formatters retrieval.

  async getUplinkDecoder(appId, version) {
    const result = await this._api.GetUplinkDecoder({
      routeParams: {
        'application_ids.application_id': appId,
        'version_ids.brand_id': version.brand_id,
        'version_ids.model_id': version.model_id,
        'version_ids.firmware_version': version.firmware_version,
        'version_ids.band_id': version.band_id,
      },
    })

    return Marshaler.payloadSingleResponse(result)
  }

  async getDownlinkDecoder(appId, version) {
    const result = await this._api.GetDownlinkDecoder({
      routeParams: {
        'application_ids.application_id': appId,
        'version_ids.brand_id': version.brand_id,
        'version_ids.model_id': version.model_id,
        'version_ids.firmware_version': version.firmware_version,
        'version_ids.band_id': version.band_id,
      },
    })

    return Marshaler.payloadSingleResponse(result)
  }

  async getDownlinkEncoder(appId, version) {
    const result = await this._api.GetDownlinkEncoder({
      routeParams: {
        'application_ids.application_id': appId,
        'version_ids.brand_id': version.brand_id,
        'version_ids.model_id': version.model_id,
        'version_ids.firmware_version': version.firmware_version,
        'version_ids.band_id': version.band_id,
      },
    })

    return Marshaler.payloadSingleResponse(result)
  }
}

export default DeviceRepository
