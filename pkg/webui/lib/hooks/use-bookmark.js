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

import { useEffect, useState } from 'react'
import { useDispatch } from 'react-redux'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

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

const entityRequestMap = {
  application: getApplication,
  gateway: getGateway,
  organization: getOrganization,
  user: getUser,
  client: getClient,
  device: getDevice,
}

/**
 * Module for getting the title, path, and icon corresponding to a bookmark.
 * Using this because for each bookmark we need to make a request to get the name of the entity, compose the path and find the icon.
 *
 * @param {*} bookmark - The bookmark object.
 * @returns {*} - An object containing the title, path, and icon of the bookmark.
 */
const useBookmark = bookmark => {
  const dispatch = useDispatch()
  const entityIds = bookmark.entity_ids
  // Get the entity of the bookmark.
  const entity = Object.keys(entityIds)[0].replace('_ids', '')
  // Get the entity id.
  const entityId = {
    id: entityIds[`${entity}_ids`][`${entity}_id`],
  }
  if (entity === 'device') {
    entityId.appId = entityIds[`${entity}_ids`].application_ids.application_id
  }
  const [bookmarkTitle, setBookmarkTitle] = useState('')
  useEffect(() => {
    const fetchEntity = async () => {
      // TODO: Do thios witj selectors checking the selected entity and see if it's already in store.
      const response = await dispatch(attachPromise(entityRequestMap[entity](entityId.id, 'name')))
      setBookmarkTitle(response.name || entityIds[`${entity}_ids`][`${entity}_id`])
    }
    fetchEntity()
  }, [entity, entityId, dispatch, entityIds])

  // Get the icon corresponding to the entity.
  const icon = iconMap[entity]
  // Get the path corresponding to the entity.
  const path =
    entity === 'device'
      ? `/applications/${entityIds.device_ids.application_ids.application_id}/devices/${entityIds.device_ids.device_id}`
      : `/${entity}s/${entityIds[`${entity}_ids`][`${entity}_id`]}`

  return { title: bookmarkTitle, path, icon }
}

export default useBookmark
