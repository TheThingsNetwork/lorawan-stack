// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

import { mapAttributesToFormValue, mapFormValueToAttributes } from '@console/lib/attributes'

export const mapFormValuesToApplication = values => ({
  name: values.name,
  description: values.description,
  attributes: mapFormValueToAttributes(values.attributes),
  skip_payload_crypto: values.skip_payload_crypto,
})

export const mapApplicationToFormValues = application => ({
  ids: {
    application_id: application.ids.application_id,
  },
  name: application.name,
  description: application.description,
  attributes: mapAttributesToFormValue(application.attributes),
  skip_payload_crypto: application.skip_payload_crypto,
})
