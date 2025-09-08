// Copyright © 2025 The Things Network Foundation, The Things Industries B.V.
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

import autoBind from 'auto-bind'

import Marshaler from '../util/marshaler'

class NsMACSettingsProfiles {
  constructor(service) {
    this._api = service
    autoBind(this)
  }

  async getAll(applicationId, params) {
    const result = await this._api.List(
      {
        routeParams: { 'application_ids.application_id': applicationId },
      },
      {
        ...params,
      },
    )
    return Marshaler.payloadListResponse('mac_settings_profiles', result)
  }

  async get(applicationId, macSettingsProfileId) {
    const result = await this._api.Get({
      routeParams: {
        'mac_settings_profile_ids.application_ids.application_id': applicationId,
        'mac_settings_profile_ids.profile_id': macSettingsProfileId,
      },
    })
    return Marshaler.payloadSingleResponse(result)
  }

  async create(applicationId, macSettingsProfile) {
    const result = await this._api.Create(
      {
        routeParams: { 'mac_settings_profile.ids.application_ids.application_id': applicationId },
      },
      {
        macSettingsProfile,
      },
    )
    return Marshaler.payloadSingleResponse(result)
  }

  async update(applicationId, macSettingsProfileId, patch) {
    const result = await this._api.Update(
      {
        routeParams: {
          'mac_settings_profile_ids.application_ids.application_id': applicationId,
          'mac_settings_profile_ids.profile_id': macSettingsProfileId,
        },
      },
      {
        mac_settings_profile_ids: {
          application_ids: { application_id: applicationId },
          profile_id: macSettingsProfileId,
        },
        mac_settings_profile: patch,
      },
    )
    return Marshaler.payloadSingleResponse(result)
  }

  async delete(applicationId, profileId) {
    const response = await this._api.Delete({
      routeParams: {
        'mac_settings_profile_ids.application_ids.application_id': applicationId,
        'mac_settings_profile_ids.profile_id': profileId,
      },
    })

    return Marshaler.payloadSingleResponse(response)
  }
}

export default NsMACSettingsProfiles
