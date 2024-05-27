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

import { createFetchingSelector } from '@ttn-lw/lib/store/selectors/fetching'
import { createErrorSelector } from '@ttn-lw/lib/store/selectors/error'

import { GET_DEVICE_TEMPLATE_FORMATS_BASE } from '@console/store/actions/device-template-formats'

const selectDeviceTemplateFormatsStore = state => state.deviceTemplateFormats

export const selectDeviceTemplateFormats = state => {
  const store = selectDeviceTemplateFormatsStore(state)

  return store.formats
}

export const selectDeviceTemplateFormatsError = createErrorSelector(
  GET_DEVICE_TEMPLATE_FORMATS_BASE,
)
export const selectDeviceTemplateFormatsFetching = createFetchingSelector(
  GET_DEVICE_TEMPLATE_FORMATS_BASE,
)
