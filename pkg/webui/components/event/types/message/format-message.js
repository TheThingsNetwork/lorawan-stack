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

import { getEventDataType } from '..'

const formatMessageData = function(data) {
  if (!data) {
    return null
  }

  const { '@type': t, ...rest } = data
  const type = getEventDataType(t)

  switch (type) {
    case 'UplinkMessage':
    case 'DownlinkMessage':
    case 'ApplicationUp':
      return rest
    default:
      return null
  }
}

export { formatMessageData as default }
