// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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
import { unstable_usePrompt } from 'react-router-dom'

import PortalledModal from '@ttn-lw/components/modal/portalled'

import PropTypes from '@ttn-lw/lib/prop-types'

const Prompt = props => {
  const { modal, children, message, when } = props
  const [showModal, setShowModal] = React.useState(false)

  // The usage of `unstable_usePrompt` might change as the library updates.
  const continueNavigation = unstable_usePrompt(when, message)

  const handleModalComplete = React.useCallback(
    approved => {
      setShowModal(false)
      if (approved) {
        continueNavigation()
      }
    },
    [continueNavigation],
  )

  React.useEffect(() => {
    if (when) {
      setShowModal(true)
    }
  }, [when])

  return (
    <PortalledModal visible={showModal} {...modal} approval onComplete={handleModalComplete}>
      {children}
    </PortalledModal>
  )
}

Prompt.propTypes = {
  children: PropTypes.oneOfType([PropTypes.arrayOf(PropTypes.node), PropTypes.node]),
  message: PropTypes.string.isRequired,
  modal: PropTypes.shape({ ...PortalledModal.Modal.propTypes }).isRequired,
  when: PropTypes.bool.isRequired,
}

Prompt.defaultProps = {
  children: undefined,
}

export default Prompt
