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

import createRequestActions from '@ttn-lw/lib/store/actions/create-request-actions'
import { createErrorSelector } from '@ttn-lw/lib/store/selectors/error'
import { createFetchingSelector } from '@ttn-lw/lib/store/selectors/fetching'

export const GET_DEVICE_TEMPLATE_FORMATS_BASE = 'GET_DEVICE_TEMPLATE_FORMATS'
export const [
  {
    request: GET_DEVICE_TEMPLATE_FORMATS,
    success: GET_DEVICE_TEMPLATE_FORMATS_SUCCESS,
    failure: GET_DEVICE_TEMPLATE_FORMATS_FAILURE,
  },
  {
    request: getDeviceTemplateFormats,
    success: getDeviceTemplateFormatsSuccess,
    failure: getDeviceTemplateFormatsFailure,
  },
] = createRequestActions(GET_DEVICE_TEMPLATE_FORMATS_BASE)
export const getDeviceTemplateFormatsFetching = createFetchingSelector(
  GET_DEVICE_TEMPLATE_FORMATS_BASE,
)
export const getDeviceTemplateFormatsError = createErrorSelector(GET_DEVICE_TEMPLATE_FORMATS_BASE)

export const CONVERT_TEMPLATE_BASE = 'CONVERT_TEMPLATE'
export const [
  {
    request: CONVERT_TEMPLATE,
    success: CONVERT_TEMPLATE_SUCCESS,
    failure: CONVERT_TEMPLATE_FAILURE,
  },
  { request: convertTemplate, success: convertTemplateSuccess, failure: convertTemplateFailure },
] = createRequestActions(CONVERT_TEMPLATE_BASE, (format_id, data) => ({ format_id, data }))
