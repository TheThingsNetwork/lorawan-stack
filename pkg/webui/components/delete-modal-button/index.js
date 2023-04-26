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

import Input from '@ttn-lw/components/input'
import ModalButton from '@ttn-lw/components/button/modal-button'
import Form from '@ttn-lw/components/form'
import Checkbox from '@ttn-lw/components/checkbox'
import Notification from '@ttn-lw/components/notification'
import Link from '@ttn-lw/components/link'

import Message from '@ttn-lw/lib/components/message'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import style from './delete-modal-button.styl'

const DeleteModalButton = props => {
  const {
    confirmMessage,
    defaultMessage,
    entityId,
    entityName,
    mayPurge,
    message,
    onApprove,
    onCancel,
    purgeMessage,
    shouldConfirm,
    title,
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
      onApprove={handleDeleteApprove}
      onCancel={onCancel}
      message={message}
      modalData={{
        title: sharedMessages.deleteModalConfirmDeletion,
        approveButtonProps: {
          disabled: shouldConfirm && confirmId !== entityId,
          icon: 'delete',
          message,
        },
        children: (
          <div>
            <Message
              content={title}
              values={{ entityName: name, pre: name => <pre className={style.id}>{name}</pre> }}
              component="span"
            />
            <Message
              content={purgeEntity ? purgeMessage : defaultMessage}
              values={{ strong: txt => <strong>{txt}</strong> }}
              component="p"
            />
            <Form initialValues={initialValues}>
              {mayPurge && (
                <Form.Field
                  name="purge"
                  className={style.hardDeleteCheckbox}
                  component={Checkbox}
                  onChange={handlePurgeEntityChange}
                  title={sharedMessages.deleteModalReleaseIdTitle}
                  label={sharedMessages.deleteModalReleaseIdLabel}
                />
              )}
              {purgeEntity && (
                <Notification
                  small
                  warning
                  content={sharedMessages.deleteModalPurgeWarning}
                  messageValues={{
                    strong: txt => <strong>{txt}</strong>,
                    DocLink: txt => (
                      <Link.DocLink primary raw path="/the-things-stack/management/purge/">
                        {txt}
                      </Link.DocLink>
                    ),
                  }}
                />
              )}
              {shouldConfirm && (
                <>
                  <Message
                    content={confirmMessage}
                    values={{ entityId, pre: id => <pre className={style.id}>{id}</pre> }}
                    component="span"
                  />
                  <Input
                    className={style.confirmInput}
                    data-test-id="confirm_deletion"
                    value={confirmId}
                    onChange={setConfirmId}
                  />
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
  confirmMessage: PropTypes.message,
  defaultMessage: PropTypes.message,
  entityId: PropTypes.string.isRequired,
  entityName: PropTypes.string,
  mayPurge: PropTypes.bool,
  message: PropTypes.message.isRequired,
  onApprove: PropTypes.func,
  onCancel: PropTypes.func,
  onlyPurge: PropTypes.bool,
  purgeMessage: PropTypes.message,
  shouldConfirm: PropTypes.bool,
  title: PropTypes.message,
}

DeleteModalButton.defaultProps = {
  entityName: undefined,
  onApprove: undefined,
  onCancel: undefined,
  shouldConfirm: false,
  mayPurge: false,
  defaultMessage: sharedMessages.deleteModalDefaultMessage,
  purgeMessage: sharedMessages.deleteModalPurgeMessage,
  title: sharedMessages.deleteModalTitle,
  confirmMessage: sharedMessages.deleteModalConfirmMessage,
  onlyPurge: false,
}

export default DeleteModalButton
