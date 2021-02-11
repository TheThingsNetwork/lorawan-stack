// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

import { merge } from 'lodash'

import { natsUrl as natsUrlRegexp } from '@console/lib/regexp'

import providers from './providers'

const natsBlankValues = {
  username: '',
  password: '',
  address: '',
  port: '',
  secure: false,
  _use_credentials: true,
}

const mqttBlankValues = {
  server_url: '',
  client_id: '',
  username: '',
  password: '',
  subscribe_qos: '',
  publish_qos: '',
  use_tls: false,
  tls_ca: '',
  tls_client_cert: '',
  tls_client_key: '',
  _use_credentials: true,
}

export const mapNatsFormValues = nats => {
  try {
    const res = nats.server_url.match(natsUrlRegexp)
    return {
      secure: res[2] === 'tls',
      username: res[5],
      password: res[7],
      address: res[8],
      port: res[10],
      _use_credentials: Boolean(res[5] || res[7]),
    }
  } catch {
    return {}
  }
}

export const mapMqttFormValues = mqtt =>
  merge({}, mqttBlankValues, mqtt, {
    _use_credentials: Boolean(mqtt.username || mqtt.password),
  })

const mapPubsubMessageTypeToFormValue = messageType =>
  (messageType && { enabled: true, value: messageType.topic }) || { enabled: false, value: '' }

export const mapPubsubToFormValues = pubsub => {
  const isNats = 'nats' in pubsub
  const isMqtt = 'mqtt' in pubsub
  let provider = blankValues._provider
  if (isNats) {
    provider = providers.NATS
  } else if (isMqtt) {
    provider = providers.MQTT
  }
  const result = {
    pub_sub_id: pubsub.ids.pub_sub_id,
    base_topic: pubsub.base_topic,
    format: pubsub.format,
    _provider: provider,
    nats: isNats ? mapNatsFormValues(pubsub.nats) : natsBlankValues,
    mqtt: isMqtt ? mapMqttFormValues(pubsub.mqtt) : mqttBlankValues,
    downlink_ack: mapPubsubMessageTypeToFormValue(pubsub.downlink_ack),
    downlink_failed: mapPubsubMessageTypeToFormValue(pubsub.downlink_failed),
    downlink_nack: mapPubsubMessageTypeToFormValue(pubsub.downlink_nack),
    downlink_push: mapPubsubMessageTypeToFormValue(pubsub.downlink_push),
    downlink_queued: mapPubsubMessageTypeToFormValue(pubsub.downlink_queued),
    downlink_queue_invalidated: mapPubsubMessageTypeToFormValue(pubsub.downlink_queue_invalidated),
    downlink_replace: mapPubsubMessageTypeToFormValue(pubsub.downlink_replace),
    downlink_sent: mapPubsubMessageTypeToFormValue(pubsub.downlink_sent),
    join_accept: mapPubsubMessageTypeToFormValue(pubsub.join_accept),
    location_solved: mapPubsubMessageTypeToFormValue(pubsub.location_solved),
    service_data: mapPubsubMessageTypeToFormValue(pubsub.service_data),
    uplink_message: mapPubsubMessageTypeToFormValue(pubsub.uplink_message),
  }

  return result
}

const mapNatsConfigFormValueToNatsServerUrl = ({
  username,
  password,
  address,
  port,
  secure,
  _use_credentials,
}) =>
  `${secure ? 'tls' : 'nats'}://${
    _use_credentials ? `${username}:${password}@` : ''
  }${address}:${port}`

const mapMessageTypeFormValueToPubsubMessageType = formValue =>
  (formValue.enabled && { topic: formValue.value }) || null

export const mapFormValuesToPubsub = (values, appId) => {
  const result = {
    ids: {
      application_ids: {
        application_id: appId,
      },
      pub_sub_id: values.pub_sub_id,
    },
    base_topic: values.base_topic,
    format: values.format,
    downlink_ack: mapMessageTypeFormValueToPubsubMessageType(values.downlink_ack),
    downlink_failed: mapMessageTypeFormValueToPubsubMessageType(values.downlink_failed),
    downlink_nack: mapMessageTypeFormValueToPubsubMessageType(values.downlink_nack),
    downlink_push: mapMessageTypeFormValueToPubsubMessageType(values.downlink_push),
    downlink_queued: mapMessageTypeFormValueToPubsubMessageType(values.downlink_queued),
    downlink_queue_invalidated: mapMessageTypeFormValueToPubsubMessageType(
      values.downlink_queue_invalidated,
    ),
    downlink_replace: mapMessageTypeFormValueToPubsubMessageType(values.downlink_replace),
    downlink_sent: mapMessageTypeFormValueToPubsubMessageType(values.downlink_sent),
    join_accept: mapMessageTypeFormValueToPubsubMessageType(values.join_accept),
    location_solved: mapMessageTypeFormValueToPubsubMessageType(values.location_solved),
    service_data: mapMessageTypeFormValueToPubsubMessageType(values.service_data),
    uplink_message: mapMessageTypeFormValueToPubsubMessageType(values.uplink_message),
  }

  switch (values._provider) {
    case providers.NATS:
      result.nats = {
        server_url: mapNatsConfigFormValueToNatsServerUrl(values.nats),
      }
      break
    case providers.MQTT:
      result.mqtt = values.mqtt
      delete result.mqtt._use_credentials
      break
  }
  return result
}

export const blankValues = {
  pub_sub_id: '',
  base_topic: '',
  format: '',
  _provider: providers.NATS,
  nats: natsBlankValues,
  mqtt: mqttBlankValues,
  downlink_ack: { enabled: false, value: '' },
  downlink_failed: { enabled: false, value: '' },
  downlink_nack: { enabled: false, value: '' },
  downlink_push: { enabled: false, value: '' },
  downlink_queued: { enabled: false, value: '' },
  downlink_queue_invalidated: { enabled: false, value: '' },
  downlink_replace: { enabled: false, value: '' },
  downlink_sent: { enabled: false, value: '' },
  join_accept: { enabled: false, value: '' },
  location_solved: { enabled: false, value: '' },
  service_data: { enabled: false, value: '' },
  uplink_message: { enabled: false, value: '' },
}
