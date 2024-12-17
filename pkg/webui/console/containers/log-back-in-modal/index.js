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

import React from 'react'
import { defineMessages } from 'react-intl'

import { IconRefresh } from '@ttn-lw/components/icon'
import Modal from '@ttn-lw/components/modal'

const m = defineMessages({
  modalTitle: 'Please sign in again',
  modalMessage:
    "You were signed out of the Console. You can press 'Reload' to log back into the Console again.",
  buttonMessage: 'Reload',
})

const reload = () => {
  window.location.reload()
}

const LogBackInModal = () => (
  <Modal
    approval={false}
    buttonMessage={m.buttonMessage}
    message={m.modalMessage}
    title={m.modalTitle}
    onComplete={reload}
    approveButtonProps={{
      icon: IconRefresh,
      primary: true,
    }}
  />
)

export default LogBackInModal
