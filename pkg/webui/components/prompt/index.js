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
import { Prompt as RouterPrompt } from 'react-router-dom'

import PortalledModal from '@ttn-lw/components/modal/portalled'

import PropTypes from '@ttn-lw/lib/prop-types'

/*
 * `<Prompt />` is used to prompt the user before navigating from the current page. This is
 * helpful to avoid losing the state of the current page because of accidental misclick, for example,
 * for half-filled forms.
 */
const Prompt = props => {
  const { modal, children, when, shouldBlockNavigation, onApprove, onCancel } = props

  const [state, setState] = React.useState({
    showModal: false,
    nextLocation: undefined,
    confirmedLocationChange: false,
  })
  const { showModal, nextLocation, confirmedLocationChange } = state

  const handleModalShow = React.useCallback(nextLocation => {
    setState(prev => ({ ...prev, showModal: true, nextLocation }))
  }, [])

  const handleModalHide = React.useCallback(() => {
    setState(prev => ({ ...prev, showModal: false }))
  }, [])

  const handleModalComplete = React.useCallback(
    approved => {
      setState(prev => ({ ...prev, confirmedLocationChange: approved }))
      handleModalHide()
    },
    [handleModalHide],
  )

  const handlePromptTrigger = React.useCallback(
    location => {
      if (!confirmedLocationChange && shouldBlockNavigation(location)) {
        handleModalShow(location)

        return false
      }

      return true
    },
    [handleModalShow, shouldBlockNavigation, confirmedLocationChange],
  )

  React.useEffect(() => {
    if (confirmedLocationChange) {
      onApprove(nextLocation)
    } else {
      onCancel(nextLocation)
    }
  }, [confirmedLocationChange, nextLocation, onApprove, onCancel])

  return (
    <>
      <RouterPrompt when={when} message={handlePromptTrigger} />
      <PortalledModal visible={showModal} approval onComplete={handleModalComplete} modal={modal}>
        {children}
      </PortalledModal>
    </>
  )
}

Prompt.propTypes = {
  children: PropTypes.oneOfType([PropTypes.arrayOf(PropTypes.node), PropTypes.node]).isRequired,
  modal: PropTypes.shape({ ...PortalledModal.Modal.propTypes }).isRequired,
  onApprove: PropTypes.func.isRequired,
  onCancel: PropTypes.func.isRequired,
  shouldBlockNavigation: PropTypes.func.isRequired,
  when: PropTypes.bool.isRequired,
}

export default Prompt
