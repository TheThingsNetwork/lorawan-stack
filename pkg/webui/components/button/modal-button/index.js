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

import React, { useCallback } from 'react'

import PortalledModal from '@ttn-lw/components/modal/portalled'

import PropTypes from '@ttn-lw/lib/prop-types'

import Button from '..'

// ModalButton is a button which needs a modal confirmation to complete the
// action. It can be used as an easy way to get the users explicit confirmation
// before doing an action, e.g. Deleting a resource.
const ModalButton = ({ modalData, message, onApprove, onCancel, ...rest }) => {
  const [modalVisible, setModalVisible] = React.useState(false)

  const handleClick = useCallback(() => {
    if (!modalData) {
      // No modal data likely means a faulty implementation, so since it's
      // likely best to not do anything in this case
      return
    }

    setModalVisible(true)
  }, [modalData])

  const handleComplete = useCallback(
    confirmed => {
      if (confirmed) {
        onApprove()
      } else {
        onCancel()
      }
      setModalVisible(false)
    },
    [onCancel, onApprove],
  )

  const modalComposedData = {
    approval: true,
    danger: true,
    buttonMessage: message,
    title: message,
    onComplete: handleComplete,
    ...modalData,
  }

  return (
    <React.Fragment>
      <PortalledModal visible={modalVisible} {...modalComposedData} />
      <Button onClick={handleClick} message={message} {...rest} />
    </React.Fragment>
  )
}

ModalButton.defaultProps = {
  onApprove: () => null,
  onCancel: () => null,
}

ModalButton.propTypes = {
  message: PropTypes.message.isRequired,
  modalData: PropTypes.shape({ ...PortalledModal.Modal.propTypes }).isRequired,
  onApprove: PropTypes.func,
  onCancel: PropTypes.func,
}

export default ModalButton
