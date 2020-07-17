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

import Yup from '@ttn-lw/lib/yup'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import getHostFromUrl from '@ttn-lw/lib/host-from-url'

import { address as addressRegexp } from '@console/lib/regexp'
import { ACTIVATION_MODES } from '@console/lib/device-utils'

const toUndefined = value => (!Boolean(value) ? undefined : value)

const validationSchema = Yup.object()
  .shape({
    application_server_address: Yup.string().when(
      ['$asUrl', '$asEnabled', '_activation_mode'],
      (asUrl, asEnabled, activationMode, schema) => {
        if (activationMode === ACTIVATION_MODES.NONE) {
          return schema.strip()
        }

        if (!asEnabled) {
          return schema.matches(addressRegexp, sharedMessages.validateAddressFormat)
        }

        return schema
          .matches(addressRegexp, sharedMessages.validateAddressFormat)
          .default(getHostFromUrl(asUrl))
      },
    ),
    network_server_address: Yup.string().when(
      ['$nsUrl', '$nsEnabled', '$mayEditKeys', '_activation_mode'],
      (nsUrl, nsEnabled, mayEditKeys, activationMode, schema) => {
        if (activationMode === ACTIVATION_MODES.NONE) {
          return schema.strip()
        }

        if (!nsEnabled) {
          return schema.matches(addressRegexp, sharedMessages.validateAddressFormat)
        }

        if (!mayEditKeys) {
          if (activationMode === ACTIVATION_MODES.OTAA) {
            return schema
              .matches(addressRegexp, sharedMessages.validateAddressFormat)
              .default(getHostFromUrl(nsUrl))
          }

          return schema.matches(addressRegexp, sharedMessages.validateAddressFormat)
        }

        return schema
          .matches(addressRegexp, sharedMessages.validateAddressFormat)
          .default(getHostFromUrl(nsUrl))
      },
    ),
    join_server_address: Yup.string().when(
      ['_activation_mode', '$jsUrl', '$jsEnabled', '_external_js'],
      (activationMode, jsUrl, jsEnabled, externalJs, schema) => {
        if (externalJs || activationMode !== ACTIVATION_MODES.OTAA || !jsEnabled) {
          return schema.strip()
        }

        return schema
          .matches(addressRegexp, sharedMessages.validateAddressFormat)
          .transform(toUndefined)
          .default(getHostFromUrl(jsUrl))
      },
    ),
    lorawan_version: Yup.string().when(['_activation_mode'], (activationMode, schema) => {
      if (activationMode === ACTIVATION_MODES.NONE) {
        return schema.strip()
      }

      return schema.required(sharedMessages.validateRequired)
    }),
    supports_join: Yup.boolean()
      .transform(() => undefined)
      .when(['$jsEnabled', '_activation_mode'], (jsEnabled, activationMode, schema) => {
        if (!jsEnabled || activationMode === ACTIVATION_MODES.NONE) {
          return schema.strip()
        }

        if (
          activationMode === ACTIVATION_MODES.ABP ||
          activationMode === ACTIVATION_MODES.MULTICAST
        ) {
          return schema.default(false)
        }

        if (activationMode === ACTIVATION_MODES.OTAA) {
          return schema.default(true)
        }

        return schema
      }),
    multicast: Yup.boolean()
      .transform(() => undefined)
      .when(['$nsEnabled', '_activation_mode'], (nsEnabled, activationMode, schema) => {
        if (!nsEnabled || activationMode === ACTIVATION_MODES.NONE) {
          return schema.strip()
        }

        if (activationMode === ACTIVATION_MODES.OTAA || activationMode === ACTIVATION_MODES.ABP) {
          return schema.default(false)
        }

        if (activationMode === ACTIVATION_MODES.MULTICAST) {
          return schema.default(true)
        }

        return schema
      }),
    _external_js: Yup.boolean(),
    _activation_mode: Yup.mixed().when(
      ['$nsEnabled', '$jsEnabled', '$mayEditKeys'],
      (nsEnabled, jsEnabled, mayEditKeys, schema) => {
        const canCreateNs = nsEnabled && mayEditKeys
        const canCreateJs = jsEnabled

        if (!canCreateJs && !canCreateNs) {
          return schema.oneOf([ACTIVATION_MODES.NONE]).required(sharedMessages.validateRequired)
        }

        if (!canCreateNs) {
          return schema
            .oneOf([ACTIVATION_MODES.OTAA, ACTIVATION_MODES.NONE])
            .required(sharedMessages.validateRequired)
        }

        if (!canCreateJs) {
          return schema
            .oneOf([ACTIVATION_MODES.ABP, ACTIVATION_MODES.MULTICAST, ACTIVATION_MODES.NONE])
            .required(sharedMessages.validateRequired)
        }

        return schema
          .oneOf(Object.values(ACTIVATION_MODES))
          .required(sharedMessages.validateRequired)
      },
    ),
  })
  .noUnknown()

export default validationSchema
