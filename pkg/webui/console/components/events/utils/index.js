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

export const getEventId = event => event.unique_id

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

export const getSignalInformation = data => {
  const notFound = { snr: NaN, rssi: NaN }
  if (!data) {
    return notFound
  }
  const { rx_metadata } = data
  if (!rx_metadata || rx_metadata.length === 0) {
    return notFound
  }
  const { snr, rssi } = rx_metadata.reduce((prev, current) =>
    prev.snr >= current.snr ? prev : current,
  )
  return { snr, rssi }
}

export const getDataRate = data => {
  if (!data) {
    return undefined
  }
  const { settings } = data
  if (!settings) {
    return undefined
  }
  const { data_rate } = settings
  if (!data_rate) {
    return undefined
  }
  const { lora, fsk, lrfhss } = data_rate
  // The encoding below mimics the encoding of the `modu` field of the UDP packet forwarder.
  if (lora) {
    const { bandwidth, spreading_factor } = lora
    return `SF${spreading_factor}BW${bandwidth / 1000}`
  } else if (fsk) {
    const { bit_rate } = fsk
    return `${bit_rate}`
  } else if (lrfhss) {
    const { modulation_type, operating_channel_width } = lrfhss
    return `M${modulation_type ?? 0}CW${operating_channel_width / 1000}`
  }
  return undefined
}
