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
import Form from '@ttn-lw/components/form'
import Checkbox from '@ttn-lw/components/checkbox'
import Notification from '@ttn-lw/components/notification'
import Link from '@ttn-lw/components/link'

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
  releaseIdTitle: 'Entity purge (admin only)',
  releaseIdLabel: 'Also release entity IDs (purge)',
  purgeWarning:
    'Releasing the entity IDs will make it possible to register a new entity with the same ID. Note that this <strong>irreversible</strong> and may lead to <strong>other users gaining access to historical data of the entity if they register an entity with the same ID</strong> . Please make sure you understand the implications of purging as described <DocLink>here</DocLink>.',
})

const DeleteModalButton = props => {
  const {
    entityId,
    entityName,
    onApprove,
    onCancel,
    shouldConfirm,
    mayPurge,
    message,
    onlyPurge,
  } = props

  const name = entityName ? entityName : entityId

  const [confirmId, setConfirmId] = React.useState('')
  const [purgeEntity, setPurgeEntity] = React.useState(onlyPurge)
  const handlePurgeEntityChange = React.useCallback(() => {
    setPurgeEntity(purge => !purge)
    setConfirmId('')
  }, [])

  const handleDeleteApprove = React.useCallback(() => {
    onApprove(purgeEntity)
  }, [onApprove, purgeEntity])

  const initialValues = React.useMemo(
    () => ({
      purge: false,
    }),
    [],
  )

  return (
    <ModalButton
      type="button"
      icon="delete"
      danger
      naked
      onApprove={handleDeleteApprove}
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
              content={purgeEntity ? m.modalPurgeWarning : m.modalDefaultWarning}
              values={{ strong: txt => <strong>{txt}</strong> }}
              component="p"
            />
            <Form initialValues={initialValues}>
              {(mayPurge || !onlyPurge) && (
                <Form.Field
                  name="purge"
                  className={style.hardDeleteCheckbox}
                  component={Checkbox}
                  onChange={handlePurgeEntityChange}
                  title={m.releaseIdTitle}
                  label={m.releaseIdLabel}
                />
              )}
              {purgeEntity && (
                <Notification
                  small
                  warning
                  content={m.purgeWarning}
                  messageValues={{
                    strong: txt => <strong>{txt}</strong>,
                    DocLink: txt => (
                      <Link.DocLink primary raw path="/reference/purge">
                        {txt}
                      </Link.DocLink>
                    ),
                  }}
                />
              )}
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
            </Form>
          </div>
        ),
      }}
    />
  )
}

DeleteModalButton.propTypes = {
  entityId: PropTypes.string.isRequired,
  entityName: PropTypes.string,
  mayPurge: PropTypes.bool,
  message: PropTypes.message.isRequired,
  onApprove: PropTypes.func,
  onCancel: PropTypes.func,
  onlyPurge: PropTypes.bool,
  shouldConfirm: PropTypes.bool,
}

DeleteModalButton.defaultProps = {
  entityName: undefined,
  onApprove: undefined,
  onCancel: undefined,
  shouldConfirm: false,
  mayPurge: false,
  onlyPurge: false,
}

export default DeleteModalButton
