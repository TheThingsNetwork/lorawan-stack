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

const mapWebhookMessageTypeToFormValue = messageType =>
  (messageType && { enabled: true, value: messageType.path }) || { enabled: false, value: '' }

const mapWebhookHeadersTypeToFormValue = headersType =>
  (headersType &&
    Object.keys(headersType).reduce(
      (result, key) =>
        result.concat({
          key,
          value: headersType[key],
        }),
      [],
    )) ||
  []

export const mapWebhookToFormValues = webhook => ({
  webhook_id: webhook.ids.webhook_id,
  base_url: webhook.base_url,
  format: webhook.format,
  headers: mapWebhookHeadersTypeToFormValue(webhook.headers),
  uplink_message: mapWebhookMessageTypeToFormValue(webhook.uplink_message),
  join_accept: mapWebhookMessageTypeToFormValue(webhook.join_accept),
  downlink_ack: mapWebhookMessageTypeToFormValue(webhook.downlink_ack),
  downlink_nack: mapWebhookMessageTypeToFormValue(webhook.downlink_nack),
  downlink_sent: mapWebhookMessageTypeToFormValue(webhook.downlink_sent),
  downlink_failed: mapWebhookMessageTypeToFormValue(webhook.downlink_failed),
  downlink_queued: mapWebhookMessageTypeToFormValue(webhook.downlink_queued),
  location_solved: mapWebhookMessageTypeToFormValue(webhook.location_solved),
})

const mapMessageTypeFormValueToWebhookMessageType = formValue =>
  (formValue.enabled && { path: formValue.value }) || null

const mapHeadersTypeFormValueToWebhookHeadersType = formValue =>
  (formValue &&
    formValue.reduce(
      (result, { key, value }) => ({
        ...result,
        [key]: value,
      }),
      {},
    )) ||
  null

export const mapFormValuesToWebhook = function(values, appId) {
  return {
    ids: {
      application_ids: {
        application_id: appId,
      },
      webhook_id: values.webhook_id,
    },
    base_url: values.base_url,
    format: values.format,
    headers: mapHeadersTypeFormValueToWebhookHeadersType(values.headers),
    uplink_message: mapMessageTypeFormValueToWebhookMessageType(values.uplink_message),
    join_accept: mapMessageTypeFormValueToWebhookMessageType(values.join_accept),
    downlink_ack: mapMessageTypeFormValueToWebhookMessageType(values.downlink_ack),
    downlink_nack: mapMessageTypeFormValueToWebhookMessageType(values.downlink_nack),
    downlink_sent: mapMessageTypeFormValueToWebhookMessageType(values.downlink_sent),
    downlink_failed: mapMessageTypeFormValueToWebhookMessageType(values.downlink_failed),
    downlink_queued: mapMessageTypeFormValueToWebhookMessageType(values.downlink_queued),
    location_solved: mapMessageTypeFormValueToWebhookMessageType(values.location_solved),
  }
}

export const blankValues = {
  webhook_id: undefined,
  base_url: undefined,
  format: undefined,
  uplink_message: { enabled: false, value: '' },
  join_accept: { enabled: false, value: '' },
  downlink_ack: { enabled: false, value: '' },
  downlink_nack: { enabled: false, value: '' },
  downlink_sent: { enabled: false, value: '' },
  downlink_failed: { enabled: false, value: '' },
  downlink_queued: { enabled: false, value: '' },
  location_solved: { enabled: false, value: '' },
}
