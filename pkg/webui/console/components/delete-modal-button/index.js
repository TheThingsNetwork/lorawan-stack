// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

import Input from '@ttn-lw/components/input'
import ModalButton from '@ttn-lw/components/button/modal-button'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './delete-modal-button.styl'

const m = defineMessages({
  modalTitle: 'Are you sure you want to delete <pre>{entityName}</pre>? ',
  modalDefaultWarning:
    'This will <strong>PERMANENTLY DELETE THE ENTITY ITSELF AND ALL ASSOCIATED ENTITIES</strong>, including collaborator associations. It will also <strong>NOT BE POSSIBLE TO REUSE THE ENTITY ID</strong>.',
  modalPurgeWarning:
    'This will <strong>PERMANENTLY DELETE THE ENTITY ITSELF AND ALL ASSOCIATED ENTITIES</strong>, including collaborator associations.',
  confirmMessage: 'Please enter <pre>{entityId}</pre> to confirm the deletion.',
  confirmDeletion: 'Confirm deletion',
})

const DeleteModalButton = props => {
  const { entityId, entityName, onApprove, onCancel, shouldConfirm, shouldPurge, message } = props
  const name = entityName ? entityName : entityId

  const [confirmId, setConfirmId] = React.useState('')

  return (
    <ModalButton
      type="button"
      icon="delete"
      danger
      naked
      onApprove={onApprove}
      onCancel={onCancel}
      message={message}
      modalData={{
        title: m.confirmDeletion,
        approveButtonProps: {
          disabled: shouldConfirm && confirmId !== entityId,
          icon: 'delete',
          message,
        },
        children: (
          <div>
            <Message
              content={m.modalTitle}
              values={{ entityName: name, pre: name => <pre className={style.id}>{name}</pre> }}
              component="span"
            />
            <Message
              content={shouldPurge ? m.modalPurgeWarning : m.modalDefaultWarning}
              values={{ strong: txt => <strong>{txt}</strong> }}
              component="p"
            />
            {shouldConfirm && (
              <>
                <Message
                  content={m.confirmMessage}
                  values={{ entityId, pre: id => <pre className={style.id}>{id}</pre> }}
                  component="span"
                />
                <Input className={style.confirmInput} value={confirmId} onChange={setConfirmId} />
              </>
            )}
          </div>
        ),
      }}
    />
  )
}

DeleteModalButton.propTypes = {
  entityId: PropTypes.string.isRequired,
  entityName: PropTypes.string,
  onApprove: PropTypes.func,
  onCancel: PropTypes.func,
  shouldConfirm: PropTypes.bool,
  shouldPurge: PropTypes.bool,
  message: PropTypes.message.isRequired,
}

DeleteModalButton.defaultProps = {
  entityName: undefined,
  onApprove: undefined,
  onCancel: undefined,
  shouldConfirm: false,
  shouldPurge: false,
}

export default DeleteModalButton
