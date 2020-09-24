// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

import { getDeviceId, getApplicationId } from '@ttn-lw/lib/selectors/id'

import DefaultPreview from '../previews/default'
import DefaultSyntheticEventPreview from '../previews/synthetic/default'
import SyntheticErrorEventPreview from '../previews/synthetic/error'

import { eventIconMap, dataTypeMap, applicationUpMessages } from './definitions'

export const getEventId = event => `${event.time}-${event.name}`

export const getEventIconByName = eventName => {
  const definition = eventIconMap.find(e => e.test.test(eventName))
  return definition ? definition.icon : 'event'
}

export const getPreviewComponentByDataType = dataType => {
  if (!dataType) {
    return DefaultPreview
  }

  const entries = dataType.split('.')
  const messageType = entries[entries.length - 1]

  return messageType in dataTypeMap ? dataTypeMap[messageType] : DefaultPreview
}

export const getSyntheticPreviewComponent = event => {
  if (event.isError) {
    return SyntheticErrorEventPreview
  }

  return DefaultSyntheticEventPreview
}

export const getPreviewComponent = event => {
  if (event.isSynthetic) {
    return getSyntheticPreviewComponent(event)
  } else if ('data' in event) {
    return getPreviewComponentByDataType(event.data['@type'])
  }

  return DefaultPreview
}

export const getEntityId = eventIdentifier =>
  getDeviceId(eventIdentifier) || getApplicationId(eventIdentifier)

export const getApplicationUpMessage = data =>
  Object.keys(data).find(e => applicationUpMessages.includes(e))

export const getPreviewComponentByApplicationUpMessage = message => {
  let messageType
  switch (message) {
    case 'uplink_message':
      messageType = 'ApplicationUplink'
      break
    case 'join_accept':
      messageType = 'ApplicationJoinAccept'
      break
    case 'downlink_ack':
    case 'downlink_nack':
    case 'downlink_sent':
    case 'downlink_queued':
      messageType = 'ApplicationDownlink'
      break
    case 'downlink_failed':
    case 'downlink_queue_invalidated':
      messageType = 'ApplicationInvalidatedDownlinks'
      break
    case 'location_solved':
      messageType = 'ApplicationLocation'
      break
    case 'service_data':
      messageType = 'ApplicationServiceData'
  }

  return messageType in dataTypeMap ? dataTypeMap[messageType] : DefaultPreview
}
