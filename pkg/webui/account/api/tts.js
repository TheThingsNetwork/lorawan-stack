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

import TTS, { STACK_COMPONENTS_MAP, AUTHORIZATION_MODES } from 'ttn-lw'

import toast from '@ttn-lw/components/toast'

import { selectStackConfig, selectCSRFToken } from '@ttn-lw/lib/selectors/env'

const stackConfig = selectStackConfig()
const csrfToken = selectCSRFToken()

const stack = {
  [STACK_COMPONENTS_MAP.is]: stackConfig.is.enabled ? stackConfig.is.base_url : undefined,
}

const tts = new TTS({
  authorization: {
    mode: AUTHORIZATION_MODES.SESSION,
    csrfToken,
  },
  stackConfig: stack,
  connectionType: 'http',
  proxy: false,
})

// Forward header warnings to the toast message queue.
tts.subscribe('warning', payload => {
  toast({
    title: 'Warning',
    type: toast.types.WARNING,
    message: payload,
    preventConsecutive: true,
  })
})

export default tts
