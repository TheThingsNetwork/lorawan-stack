// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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
import { useNavigate } from 'react-router-dom'
import { useDispatch, useSelector } from 'react-redux'

import { IconTrash } from '@ttn-lw/components/icon'
import PortalledModal from '@ttn-lw/components/modal/portalled'
import toast from '@ttn-lw/components/toast'
import Form from '@ttn-lw/components/form'
import Checkbox from '@ttn-lw/components/checkbox'
import Notification from '@ttn-lw/components/notification'
import Link from '@ttn-lw/components/link'
import Input from '@ttn-lw/components/input'

import RequireRequest from '@ttn-lw/lib/components/require-request'
import Message from '@ttn-lw/lib/components/message'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import { getCollaboratorsList } from '@ttn-lw/lib/store/actions/collaborators'
import { selectCollaboratorsTotalCount } from '@ttn-lw/lib/store/selectors/collaborators'

import { checkFromState, mayPurgeEntities } from '@console/lib/feature-checks'

import { getApiKeysList } from '@console/store/actions/api-keys'
import { getIsConfiguration } from '@console/store/actions/identity-server'

import { selectApiKeysTotalCount } from '@console/store/selectors/api-keys'

const DeleteEntityHeaderModal = props => {
  const {
    entity,
    entityId,
    entityName,
    visible,
    setVisible,
    setError,
    mayDeleteEntitySelector,
    mayViewOrEditEntityCollaborators,
    mayViewOrEditEntityApiKeys,
    path,
    deleteEntity,
    deleteMessage,
    deletedMessage,
    deletedErrorMessage,
    additionalConditions,
  } = props

  const [confirmId, setConfirmId] = React.useState('')
  const [purgeEntity, setPurgeEntity] = React.useState(false)
  const dispatch = useDispatch()
  const navigate = useNavigate()

  const mayPurgeEntity = useSelector(state => checkFromState(mayPurgeEntities, state))
  const mayDeleteEntity = useSelector(state => checkFromState(mayDeleteEntitySelector, state))
  const apiKeysCount = useSelector(selectApiKeysTotalCount)
  const collaboratorsCount = useSelector(state =>
    selectCollaboratorsTotalCount(state, { id: entityId }),
  )
  const hasApiKeys = apiKeysCount > 0
  const hasAddedCollaborators = collaboratorsCount > 1
  const isPristine = !hasAddedCollaborators && !hasApiKeys && !additionalConditions
  const mayViewCollaborators = useSelector(state =>
    checkFromState(mayViewOrEditEntityCollaborators, state),
  )
  const mayViewApiKeys = useSelector(state => checkFromState(mayViewOrEditEntityApiKeys, state))
  const shouldConfirmDelete = !isPristine || !mayViewCollaborators || !mayViewApiKeys

  const name = entityName ? entityName : entityId

  const handlePurgeEntityChange = React.useCallback(() => {
    setPurgeEntity(purge => !purge)
    setConfirmId('')
  }, [])

  const handleComplete = useCallback(
    async confirmed => {
      if (confirmed) {
        {
          try {
            if (setError) {
              setError(undefined)
            }
            await dispatch(attachPromise(deleteEntity(entityId, { purge: purgeEntity || false })))
            navigate(path)
            toast({
              title: entityId,
              message: deletedMessage,
              type: toast.types.SUCCESS,
            })
          } catch (error) {
            if (setError) {
              setError(error)
            }
            toast({
              title: entityId,
              message: deletedErrorMessage,
              type: toast.types.ERROR,
            })
          }
        }
      }
      setVisible(false)
    },
    [
      dispatch,
      entityId,
      navigate,
      purgeEntity,
      setError,
      setVisible,
      path,
      deleteEntity,
      deletedMessage,
      deletedErrorMessage,
    ],
  )

  const loadData = useCallback(
    async dispatch => {
      if (mayDeleteEntity) {
        if (mayViewApiKeys) {
          await dispatch(attachPromise(getApiKeysList(entity, entityId)))
        }
        if (mayViewCollaborators) {
          await dispatch(attachPromise(getCollaboratorsList(entity, entityId)))
        }
      }
      dispatch(attachPromise(getIsConfiguration()))
    },
    [entityId, mayDeleteEntity, mayViewApiKeys, mayViewCollaborators, entity],
  )

  const initialValues = React.useMemo(
    () => ({
      purge: false,
    }),
    [],
  )

  return (
    <RequireRequest requestAction={loadData}>
      <PortalledModal
        danger
        visible={visible}
        onComplete={handleComplete}
        title={sharedMessages.deleteModalConfirmDeletion}
        approveButtonProps={{
          disabled: shouldConfirmDelete && confirmId !== entityId,
          icon: IconTrash,
          primary: true,
          message: deleteMessage,
        }}
      >
        <div>
          <Message
            content={sharedMessages.deleteModalTitle}
            values={{ entityName: name, pre: name => <pre className="d-inline">{name}</pre> }}
            component="span"
          />
          <Message
            content={
              purgeEntity
                ? sharedMessages.deleteModalPurgeMessage
                : sharedMessages.deleteModalDefaultMessage
            }
            values={{ strong: txt => <strong>{txt}</strong> }}
            component="p"
          />
          <Form initialValues={initialValues}>
            {mayPurgeEntity && (
              <Form.Field
                name="purge"
                className="mt-ls-xxs"
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
            {shouldConfirmDelete && (
              <>
                <Message
                  content={sharedMessages.deleteModalConfirmMessage}
                  values={{ entityId, pre: id => <pre className="d-inline">{id}</pre> }}
                  component="span"
                />
                <Input
                  className="mt-ls-xxs"
                  data-test-id="confirm_deletion"
                  value={confirmId}
                  onChange={setConfirmId}
                />
              </>
            )}
          </Form>
        </div>
      </PortalledModal>
    </RequireRequest>
  )
}

DeleteEntityHeaderModal.propTypes = {
  additionalConditions: PropTypes.bool,
  deleteEntity: PropTypes.func.isRequired,
  deleteMessage: PropTypes.message.isRequired,
  deletedErrorMessage: PropTypes.message.isRequired,
  deletedMessage: PropTypes.message.isRequired,
  entity: PropTypes.string.isRequired,
  entityId: PropTypes.string.isRequired,
  entityName: PropTypes.string,
  mayDeleteEntitySelector: PropTypes.shape({}).isRequired,
  mayViewOrEditEntityApiKeys: PropTypes.shape({}).isRequired,
  mayViewOrEditEntityCollaborators: PropTypes.shape({}).isRequired,
  path: PropTypes.string.isRequired,
  setError: PropTypes.func,
  setVisible: PropTypes.func.isRequired,
  visible: PropTypes.bool.isRequired,
}

DeleteEntityHeaderModal.defaultProps = {
  additionalConditions: false,
  entityName: undefined,
  setError: undefined,
}

export default DeleteEntityHeaderModal
