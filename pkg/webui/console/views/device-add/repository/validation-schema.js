// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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
import { id as deviceIdRegexp } from '@ttn-lw/lib/regexp'
import getHostFromUrl from '@ttn-lw/lib/host-from-url'

import { parseLorawanMacVersion } from '@console/lib/device-utils'

import { REGISTRATION_TYPES } from '../utils'

const deviceIdSchema = Yup.string()
  .matches(deviceIdRegexp, Yup.passValues(sharedMessages.validateIdFormat))
  .min(2, Yup.passValues(sharedMessages.validateTooShort))
  .max(36, Yup.passValues(sharedMessages.validateTooLong))
  .required(sharedMessages.validateRequired)

const joinEUISchema = Yup.string().length(8 * 2, Yup.passValues(sharedMessages.validateLength))
const devEUISchema = Yup.string().length(8 * 2, Yup.passValues(sharedMessages.validateLength))

const validationSchema = Yup.object({
  // Form fields.
  version_ids: Yup.object({
    brand_id: Yup.string(),
    model_id: Yup.string(),
    hardware_version: Yup.string(),
    firmware_version: Yup.string(),
    band_id: Yup.string(),
  }),
  frequency_plan_id: Yup.string().required(sharedMessages.validateRequired),
  ids: Yup.object().when(['supports_join', 'lorawan_version'], (isOTAA, version, schema) => {
    if (isOTAA) {
      return schema.shape({
        device_id: deviceIdSchema,
        join_eui: joinEUISchema.required(sharedMessages.validateRequired),
        dev_eui: devEUISchema.required(sharedMessages.validateRequired),
      })
    }

    if (parseLorawanMacVersion(version) === 104) {
      return schema.shape({
        device_id: deviceIdSchema,
        dev_eui: devEUISchema.required(sharedMessages.validateRequired),
        join_eui: Yup.string().strip(),
      })
    }

    return schema.shape({
      join_eui: Yup.string().strip(),
      device_id: deviceIdSchema,
      dev_eui: Yup.lazy(value =>
        !value
          ? Yup.string().strip()
          : Yup.string().length(8 * 2, Yup.passValues(sharedMessages.validateLength)),
      ),
    })
  }),
  root_keys: Yup.object().when(
    ['supports_join', 'lorawan_version', '$mayEditKeys'],
    (isOTAA, version, mayEditKeys, schema) => {
      if (!mayEditKeys || !isOTAA) {
        return schema.strip()
      }

      const strippedSchema = Yup.object().strip()
      const keySchema = Yup.lazy(() =>
        mayEditKeys
          ? Yup.object().shape({
              key: Yup.string()
                .length(16 * 2, Yup.passValues(sharedMessages.validateLength))
                .required(sharedMessages.validateRequired),
            })
          : Yup.object().strip(),
      )

      if (parseLorawanMacVersion(version) < 110) {
        return schema.shape({
          nwk_key: strippedSchema,
          app_key: keySchema,
        })
      }

      return schema.shape({
        nwk_key: keySchema,
        app_key: keySchema,
      })
    },
  ),
  session: Yup.object().when(['lorawan_version', 'supports_join'], (version, isOTAA, schema) => {
    if (isOTAA) {
      return schema.strip()
    }

    const lwVersion = parseLorawanMacVersion(version)

    return schema.shape({
      dev_addr: Yup.string()
        .length(4 * 2, Yup.passValues(sharedMessages.validateLength))
        .required(sharedMessages.validateRequired),
      keys: Yup.object().shape({
        app_s_key: Yup.object().shape({
          key: Yup.string()
            .length(16 * 2, Yup.passValues(sharedMessages.validateLength))
            .required(sharedMessages.validateRequired),
        }),
        f_nwk_s_int_key: Yup.object({
          key: Yup.string()
            .length(16 * 2, Yup.passValues(sharedMessages.validateLength))
            .required(sharedMessages.validateRequired),
        }),
        s_nwk_s_int_key: Yup.lazy(() =>
          lwVersion >= 110
            ? Yup.object().shape({
                key: Yup.string()
                  .length(16 * 2, Yup.passValues(sharedMessages.validateLength))
                  .required(sharedMessages.validateRequired),
              })
            : Yup.object().strip(),
        ),
        nwk_s_enc_key: Yup.lazy(() =>
          lwVersion >= 110
            ? Yup.object().shape({
                key: Yup.string()
                  .length(16 * 2, Yup.passValues(sharedMessages.validateLength))
                  .required(sharedMessages.validateRequired),
              })
            : Yup.object().strip(),
        ),
      }),
    })
  }),
  // Referenced template values.
  supports_join: Yup.bool().default(false),
  lorawan_version: Yup.string(),
  // Derived.
  application_server_address: Yup.string().when(
    ['$asUrl', '$asEnabled'],
    (asUrl, asEnabled, schema) => {
      if (!asEnabled) {
        return schema.strip()
      }

      return schema.default(getHostFromUrl(asUrl))
    },
  ),
  network_server_address: Yup.string().when(
    ['$nsUrl', '$nsEnabled', '$mayEditKeys'],
    (nsUrl, nsEnabled, mayEditKeys, schema) => {
      if (!nsEnabled || !mayEditKeys) {
        return schema.strip()
      }

      return schema.default(getHostFromUrl(nsUrl))
    },
  ),
  join_server_address: Yup.string().when(['$jsUrl', '$jsEnabled'], (jsUrl, jsEnabled, schema) => {
    if (!jsEnabled) {
      return schema.strip()
    }

    return schema.default(getHostFromUrl(jsUrl))
  }),
  _registration: Yup.mixed().oneOf([REGISTRATION_TYPES.SINGLE, REGISTRATION_TYPES.MULTIPLE]),
})

const initialValues = {
  // Selection.
  version_ids: {
    brand_id: '',
    model_id: '',
    firmware_version: '',
    hardware_version: '',
    band_id: '',
  },
  // Registration.
  frequency_plan_id: '',
  ids: {
    device_id: '',
    dev_eui: '',
    join_eui: '',
  },
  root_keys: {
    app_key: { key: '' },
    nwk_key: { key: '' },
  },
  session: {
    dev_addr: '',
    keys: {
      app_s_key: {
        key: '',
      },
      f_nwk_s_int_key: {
        key: '',
      },
      s_nwk_s_int_key: {
        key: '',
      },
      nwk_s_enc_key: {
        key: '',
      },
    },
  },
  join_server_address: undefined,
  application_server_address: undefined,
  network_server_address: undefined,
  _registration: REGISTRATION_TYPES.SINGLE,
}

export { validationSchema as default, initialValues }
