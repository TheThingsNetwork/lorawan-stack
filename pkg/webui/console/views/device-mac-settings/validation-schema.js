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

import * as Yup from 'yup'

import { ACTIVATION_MODES } from '../../lib/device-utils'

export default Yup.object({
  _activation_mode: Yup.mixed().oneOf([
    ACTIVATION_MODES.ABP,
    ACTIVATION_MODES.OTAA,
    ACTIVATION_MODES.MULTICAST,
  ]),
  mac_settings: Yup.object({
    rx2_data_rate_index: Yup.object({
      value: Yup.number(),
    }),
    resets_f_cnt: Yup.boolean().when('activation_mode', {
      is: mode => mode === ACTIVATION_MODES.ABP,
      then: schema => schema.default(false),
      otherwise: schema => schema.strip(),
    }),
  }),
})
