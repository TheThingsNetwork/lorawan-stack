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

import { ingestError } from '@ttn-lw/lib/errors/utils'

import notificationMap from './constants'

export const getNotification = notificationType => notificationMap[notificationType]

const idToEntityMap = {
  application_ids: 'application',
  device_ids: 'end device',
  gateway_ids: 'gateway',
  user_ids: 'user',
  organization_ids: 'organization',
  client_ids: 'client',
}

export const validateNotification = (notificationType, ingestedBy) => {
  if (!(notificationType in notificationMap)) {
    ingestError(new Error(`Notification type "${notificationType}" does not exist`), {
      ingestedBy,
    })
    return false
  }

  return true
}

export const getEntity = entity_ids =>
  (entity_ids && idToEntityMap[Object.keys(entity_ids)[0]]) || 'entity'
