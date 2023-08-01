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

import React, { useCallback } from 'react'
import { defineMessages } from 'react-intl'

import Modal from '@ttn-lw/components/modal'

import PropTypes from '@ttn-lw/lib/prop-types'

const m = defineMessages({
  modalTitle: 'Confirm navigation',
  modalMessage:
    'Are you sure you want to leave this page? Your current changes have not been saved yet.',
})

const MoveAwayModal = ({ blocker }) => {
  const handleComplete = useCallback(
    result => {
      if (result) {
        return blocker.proceed?.()
      }
      return blocker.reset?.()
    },
    [blocker],
  )

  return (
    blocker.state === 'blocked' && (
      <Modal
        buttonMessage={m.modalTitle}
        message={m.modalMessage}
        title={m.modalTitle}
        onComplete={handleComplete}
      />
    )
  )
}

MoveAwayModal.propTypes = {
  blocker: PropTypes.shape({
    proceed: PropTypes.func,
    reset: PropTypes.func,
    state: PropTypes.string,
  }).isRequired,
}

export default MoveAwayModal
