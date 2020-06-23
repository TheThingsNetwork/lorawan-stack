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

import React from 'react'
import { defineMessages } from 'react-intl'

import Prompt from '@ttn-lw/components/prompt'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'
import useBeforeUnload from '@ttn-lw/lib/hooks/use-before-unload'

const m = defineMessages({
  modalTitle: 'Confirm navigation',
  modalMessage:
    'Are you sure you want to leave this page and abort the end device creation? There are unsaved changes that will be lost.',
  modalApproveText: 'Confirm navigation',
  modalCancelText: 'Stay on this page',
})

const DeviceWizardPrompt = props => {
  const { when, onApprove, onCancel, shouldBlockNavigation } = props

  useBeforeUnload(event => {
    if (when) {
      event.preventDefault()
    }
  })

  const handleShouldBlockNavigation = React.useCallback(
    location => {
      return shouldBlockNavigation(location)
    },
    [shouldBlockNavigation],
  )

  const modalProps = React.useMemo(() => {
    return {
      title: m.modalTitle,
      buttonMessage: m.modalApproveText,
      cancelButtonMessage: m.modalCancelText,
    }
  }, [])

  return (
    <Prompt
      modal={modalProps}
      when={when}
      onApprove={onApprove}
      onCancel={onCancel}
      shouldBlockNavigation={handleShouldBlockNavigation}
    >
      <Message content={m.modalMessage} />
    </Prompt>
  )
}

DeviceWizardPrompt.propTypes = {
  onApprove: PropTypes.func,
  onCancel: PropTypes.func,
  shouldBlockNavigation: PropTypes.func.isRequired,
  when: PropTypes.bool.isRequired,
}

DeviceWizardPrompt.defaultProps = {
  onApprove: () => null,
  onCancel: () => null,
}

DeviceWizardPrompt.displayName = 'DeviceWizardPrompt'

export default React.memo(DeviceWizardPrompt)
