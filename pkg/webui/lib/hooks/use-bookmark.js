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

import { useCallback } from 'react'

import { getApplication } from '@console/store/actions/applications'
import { getGateway } from '@console/store/actions/gateways'
import { getOrganization } from '@console/store/actions/organizations'
import { getUser } from '@account/store/actions/user'
import { getDevice } from '@console/store/actions/devices'
import { getClient } from '@account/store/actions/clients'

const iconMap = {
  application: 'application',
  gateway: 'gateway',
  organization: 'organization',
  user: 'user',
  client: 'client',
  device: 'device',
}

const getApplicationName = (appId, selectors) => getApplication(appId.id, selectors)
const getGatewayName = (gatewayId, selectors) => getGateway(gatewayId.id, selectors)
const getOrganizationName = (organizationId, selectors) =>
  getOrganization(organizationId.id, selectors)
const getUserName = (userId, selectors) => getUser(userId.id, selectors)
const getClientName = (clientId, selectors) => getClient(clientId.id, selectors)
const getDeviceName = (devId, selectors) => getDevice(devId.appId, devId.id, selectors)

const requestMap = {
  application: getApplicationName,
  gateway: getGatewayName,
  organization: getOrganizationName,
  user: getUserName,
  client: getClientName,
  device: getDeviceName,
}

const useBookmark = bookmark => {
  const entityName = Object.keys(bookmark.entity_ids)[0].replace('_ids', '')
  const icon = iconMap[entityName]
  const path =
    entityName === 'device'
      ? `/applications/${bookmark.entity_ids.device_ids.application_ids.application_id}/devices/${bookmark.entity_ids.device_ids.device_id}`
      : `/${entityName}s/${bookmark.entity_ids[`${entityName}_ids`][`${entityName}_id`]}`

  const funcArgument = {
    id: bookmark.entity_ids[`${entityName}_ids`][`${entityName}_id`],
  }
  if (entityName === 'device') {
    funcArgument.appId = bookmark.entity_ids[`${entityName}_ids`].application_ids.application_id
  }
  const getTitle = useCallback(() => {
    const entity = requestMap[entityName](funcArgument, 'name')
    return entity.name
  }, [entityName, funcArgument])

  const title = getTitle() || bookmark.entity_ids[`${entityName}_ids`][`${entityName}_id`]

  return { title, path, icon }
}

export default useBookmark
