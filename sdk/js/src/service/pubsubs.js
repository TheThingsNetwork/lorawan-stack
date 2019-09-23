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

import Marshaler from '../util/marshaler'

const remaps = [['nats', 'provider.nats'], ['mqtt', 'provider.mqtt']]

class PubSub {
  constructor(registry) {
    this._api = registry
  }

  _fillZeroValues(pubsub, paths) {
    // Add zero values that would otherwise be swallowed by the http bridge
    if (
      (paths.includes('provider.mqtt') || paths.includes('provider.mqtt.publish_qos')) &&
      'mqtt' in pubsub &&
      !('publish_qos' in pubsub.mqtt)
    ) {
      pubsub.mqtt.publish_qos = 'AT_MOST_ONCE'
    }

    if (
      (paths.includes('provider.mqtt') || paths.includes('provider.mqtt.subscribe_qos')) &&
      'mqtt' in pubsub &&
      !('subscribe_qos' in pubsub.mqtt)
    ) {
      pubsub.mqtt.subscribe_qos = 'AT_MOST_ONCE'
    }

    if (
      (paths.includes('provider.mqtt') || paths.includes('provider.mqtt.use_tls')) &&
      'mqtt' in pubsub &&
      !('use_tls' in pubsub.mqtt)
    ) {
      pubsub.mqtt.use_tls = false
    }
  }

  async getAll(appId, selector) {
    const result = await this._api.List(
      {
        routeParams: { 'application_ids.application_id': appId },
      },
      {
        ...Marshaler.selectorToFieldMask(selector),
      },
    )

    return Marshaler.payloadListResponse('pubsubs', result)
  }

  async create(
    appId,
    pubsub,
    mask = Marshaler.fieldMaskFromPatch(pubsub, this._api.SetAllowedFieldMaskPaths, remaps),
  ) {
    const result = await this._api.Set(
      {
        routeParams: {
          'pubsub.ids.application_ids.application_id': appId,
        },
      },
      {
        pubsub,
        field_mask: Marshaler.fieldMask(mask),
      },
    )

    return Marshaler.payloadSingleResponse(result)
  }

  async getById(appId, pubsubId, selector) {
    const fieldMask = Marshaler.selectorToFieldMask(selector)
    const paths = fieldMask.field_mask.paths
    const result = await this._api.Get(
      {
        routeParams: {
          'ids.application_ids.application_id': appId,
          'ids.pub_sub_id': pubsubId,
        },
      },
      fieldMask,
    )

    const pubsub = Marshaler.payloadSingleResponse(result)
    this._fillZeroValues(pubsub, paths)

    return pubsub
  }

  async updateById(
    appId,
    pubsubId,
    patch,
    mask = Marshaler.fieldMaskFromPatch(patch, this._api.SetAllowedFieldMaskPaths, remaps),
  ) {
    const result = await this._api.Set(
      {
        routeParams: {
          'pubsub.ids.application_ids.application_id': appId,
          'pubsub.ids.pub_sub_id': pubsubId,
        },
      },
      {
        pubsub: patch,
        field_mask: Marshaler.fieldMask(mask),
      },
    )

    return Marshaler.payloadSingleResponse(result)
  }

  async deleteById(appId, pubsubId) {
    const result = await this._api.Delete({
      routeParams: {
        'application_ids.application_id': appId,
        pub_sub_id: pubsubId,
      },
    })

    return Marshaler.payloadSingleResponse(result)
  }

  async getFormats() {
    const result = await this._api.GetFormats()

    return Marshaler.payloadSingleResponse(result)
  }
}

export default PubSub
