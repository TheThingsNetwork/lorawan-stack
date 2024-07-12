// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

class ConnectionProfiles {
  constructor(service) {
    this._api = service
    autoBind(this)
  }

  // WiFi profiles.
  // Organization.
  async getWifiProfilesForOrganization(organizationId, params, selector) {
    const response = await this._api.ManagedGatewayWiFiProfileConfigurationService.List(
      {
        routeParams: { 'collaborator.organization_ids.organization_id': organizationId },
      },
      {
        ...params,
        ...Marshaler.selectorToFieldMask(selector),
      },
    )

    return Marshaler.payloadListResponse('profiles', response)
  }

  async getWifiProfileForOrganization(organizationId, profileId, selector) {
    const fieldMask = Marshaler.selectorToFieldMask(selector)
    const response = await this._api.ManagedGatewayWiFiProfileConfigurationService.Get(
      {
        routeParams: {
          'collaborator.organization_ids.organization_id': organizationId,
          profile_id: profileId,
        },
      },
      fieldMask,
    )

    return Marshaler.payloadSingleResponse(response)
  }

  async createWifiProfileForOrganization(organizationId, profile) {
    const response = await this._api.ManagedGatewayWiFiProfileConfigurationService.Create(
      {
        routeParams: { 'collaborator.organization_ids.organization_id': organizationId },
      },
      { profile },
    )

    return Marshaler.payloadSingleResponse(response)
  }

  async deleteWifiProfileForOrganization(organizationId, profileId) {
    const response = await this._api.ManagedGatewayWiFiProfileConfigurationService.Delete({
      routeParams: {
        'collaborator.organization_ids.organization_id': organizationId,
        profile_id: profileId,
      },
    })

    return Marshaler.payloadSingleResponse(response)
  }

  async updateWifiProfileForOrganization(
    organizationId,
    profileId,
    patch,
    mask = Marshaler.fieldMaskFromPatch(patch),
  ) {
    const response = await this._api.ManagedGatewayWiFiProfileConfigurationService.Update(
      {
        routeParams: {
          'collaborator.organization_ids.organization_id': organizationId,
          'profile.profile_id': profileId,
        },
      },
      {
        profile: patch,
        field_mask: Marshaler.fieldMask(mask),
      },
    )
    return Marshaler.payloadSingleResponse(response)
  }

  // User.
  async getWifiProfilesForUser(userId, params, selector) {
    const response = await this._api.ManagedGatewayWiFiProfileConfigurationService.List(
      {
        routeParams: { 'collaborator.user_ids.user_id': userId },
      },
      {
        ...params,
        ...Marshaler.selectorToFieldMask(selector),
      },
    )

    return Marshaler.payloadListResponse('profiles', response)
  }

  async getWifiProfileForUser(userId, profileId, selector) {
    const fieldMask = Marshaler.selectorToFieldMask(selector)
    const response = await this._api.ManagedGatewayWiFiProfileConfigurationService.Get(
      {
        routeParams: { 'collaborator.user_ids.user_id': userId, profile_id: profileId },
      },
      fieldMask,
    )

    return Marshaler.payloadSingleResponse(response)
  }

  async createWifiProfileForUser(userId, profile) {
    const response = await this._api.ManagedGatewayWiFiProfileConfigurationService.Create(
      {
        routeParams: { 'collaborator.user_ids.user_id': userId },
      },
      { profile },
    )

    return Marshaler.payloadSingleResponse(response)
  }

  async deleteWifiProfileForUser(userId, profileId) {
    const response = await this._api.ManagedGatewayWiFiProfileConfigurationService.Delete({
      routeParams: {
        'collaborator.user_ids.user_id': userId,
        profile_id: profileId,
      },
    })

    return Marshaler.payloadSingleResponse(response)
  }

  async updateWifiProfileForUser(
    userId,
    profileId,
    patch,
    mask = Marshaler.fieldMaskFromPatch(patch),
  ) {
    const response = await this._api.ManagedGatewayWiFiProfileConfigurationService.Update(
      {
        routeParams: {
          'collaborator.user_ids.user_id': userId,
          'profile.profile_id': profileId,
        },
      },
      {
        profile: patch,
        field_mask: Marshaler.fieldMask(mask),
      },
    )
    return Marshaler.payloadSingleResponse(response)
  }

  // Ethernet profiles.
  // User.
  async getEthernetProfileForUser(userId, profileId, selector) {
    const fieldMask = Marshaler.selectorToFieldMask(selector)
    const response = await this._api.ManagedGatewayEthernetProfileConfigurationService.Get(
      {
        routeParams: { 'collaborator.user_ids.user_id': userId, profile_id: profileId },
      },
      fieldMask,
    )

    return Marshaler.payloadSingleResponse(response)
  }

  async createEthernetProfileForUser(userId, profile) {
    const response = await this._api.ManagedGatewayEthernetProfileConfigurationService.Create(
      {
        routeParams: { 'collaborator.user_ids.user_id': userId },
      },
      { profile },
    )

    return Marshaler.payloadSingleResponse(response)
  }

  async updateEthernetProfileForUser(
    userId,
    profileId,
    patch,
    mask = Marshaler.fieldMaskFromPatch(patch),
  ) {
    const response = await this._api.ManagedGatewayEthernetProfileConfigurationService.Update(
      {
        routeParams: {
          'collaborator.user_ids.user_id': userId,
          'profile.profile_id': profileId,
        },
      },
      {
        profile: patch,
        field_mask: Marshaler.fieldMask(mask),
      },
    )
    return Marshaler.payloadSingleResponse(response)
  }

  // Scan access points.
  async getAccessPoints(gatewayId, gatewayEui) {
    const response = await this._api.ManagedGatewayConfigurationService.ScanWiFiAccessPoints(
      {
        routeParams: { gateway_id: gatewayId },
      },
      { eui: gatewayEui },
    )

    return Marshaler.payloadSingleResponse(response)
  }
}

export default ConnectionProfiles
