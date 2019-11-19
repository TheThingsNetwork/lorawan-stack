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

import ModalButton from '../../../../components/button/modal-button'

import PropTypes from '../../../../lib/prop-types'

const m = defineMessages({
  deleteDevice: 'Delete End Device',
  deleteWarning:
    'Are you sure you want to delete "{deviceId}"? Deleting an end device cannot be undone!',
})

const DeleteSection = props => {
  const { onDelete, onDeleteSuccess, onDeleteFailure, device } = props
  const { name, ids } = device

  const handleDelete = React.useCallback(async () => {
    try {
      await onDelete()
      onDeleteSuccess()
    } catch (err) {
      onDeleteFailure()
    }
  }, [onDelete, onDeleteFailure, onDeleteSuccess])

  return (
    <div>
      <ModalButton
        type="button"
        icon="delete"
        message={m.deleteDevice}
        modalData={{
          message: { values: { deviceId: name || ids.device_id }, ...m.deleteWarning },
        }}
        onApprove={handleDelete}
        danger
      />
    </div>
  )
}

DeleteSection.propTypes = {
  device: PropTypes.device.isRequired,
  onDelete: PropTypes.func.isRequired,
  onDeleteFailure: PropTypes.func,
  onDeleteSuccess: PropTypes.func,
}

DeleteSection.defaultProps = {
  onDeleteSuccess: () => null,
  onDeleteFailure: () => null,
}

export default DeleteSection
