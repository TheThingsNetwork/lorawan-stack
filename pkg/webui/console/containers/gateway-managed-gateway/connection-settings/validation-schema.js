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

import {
  ethernetValidationSchema,
  wifiValidationSchema,
} from '@console/containers/gateway-managed-gateway/shared/validation-schema'

import Yup from '@ttn-lw/lib/yup'

export const validationSchema = Yup.object().shape({
  wifi_profile: Yup.object()
    .shape({
      _override: Yup.boolean().default(false),
      profile_id: Yup.string(),
    })
    .when('.profile_id', {
      is: profileId => profileId && profileId.includes('shared'),
      then: schema => schema.concat(wifiValidationSchema),
    }),
  ethernet_profile: ethernetValidationSchema,
})

export default validationSchema
