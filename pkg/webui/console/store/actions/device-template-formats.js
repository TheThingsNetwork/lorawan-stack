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

import { createRequestActions } from './lib'

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
    failure: geteviceTemplateFormatsFailure,
  },
] = createRequestActions(GET_DEVICE_TEMPLATE_FORMATS_BASE)
