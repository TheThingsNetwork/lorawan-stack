// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

import React, { useCallback, useState } from 'react'

import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import Notification from '@ttn-lw/components/notification'
import Radio from '@ttn-lw/components/radio-button'
import SubmitBar from '@ttn-lw/components/submit-bar'
import SubmitButton from '@ttn-lw/components/submit-button'
import toast from '@ttn-lw/components/toast'
import ModalButton from '@ttn-lw/components/button/modal-button'
import RightsGroup from '@ttn-lw/components/rights-group'

import Yup from '@ttn-lw/lib/yup'
import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import { userId as collaboratorIdRegexp } from '@ttn-lw/lib/regexp'

const validationSchema = Yup.object().shape({
  collaborator_id: Yup.string()
    .matches(collaboratorIdRegexp, Yup.passValues(sharedMessages.validateIdFormat))
    .required(sharedMessages.validateRequired),
  collaborator_type: Yup.string().required(sharedMessages.validateRequired),
  rights: Yup.array().min(1, sharedMessages.validateRights),
})

const CollaboratorForm = props => {
  const {
    collaborator,
    collaboratorId,
    deleteDisabled,
    rights,
    pseudoRights,
    error: passedError,
    update,
    onSubmit,
    onSubmitSuccess,
    onSubmitFailure,
    onDelete,
    onDeleteSuccess,
    onDeleteFailure,
    isCollaboratorUser,
    isCollaboratorAdmin,
    isCollaboratorCurrentUser,
  } = props

  const collaboratorType = isCollaboratorUser ? 'user' : 'organization'

  const [submitError, setSubmitError] = useState(undefined)
  const error = submitError || passedError

  const handleSubmit = useCallback(
    async (values, { resetForm, setSubmitting }) => {
      const { collaborator_id, collaborator_type, rights } = values

      const collaborator_ids = {
        [`${collaborator_type}_ids`]: {
          [`${collaborator_type}_id`]: collaborator_id,
        },
      }

      const collaborator = {
        ids: collaborator_ids,
        rights,
      }

      setSubmitError(undefined)

      try {
        await onSubmit(collaborator)

        resetForm({ values })
        onSubmitSuccess()
      } catch (error) {
        setSubmitting(false)
        setSubmitError(error)
        onSubmitFailure(error)
      }
    },
    [onSubmit, onSubmitFailure, onSubmitSuccess],
  )
  const handleDelete = useCallback(async () => {
    const collaborator_ids = {
      [`${collaboratorType}_ids`]: {
        [`${collaboratorType}_id`]: collaboratorId,
      },
    }
    const updatedCollaborator = {
      ids: collaborator_ids,
    }
    setSubmitError(undefined)

    try {
      await onDelete(updatedCollaborator)
      toast({
        message: sharedMessages.collaboratorDeleteSuccess,
        type: toast.types.SUCCESS,
      })
      onDeleteSuccess()
    } catch (error) {
      setSubmitError(error)
      onDeleteFailure(error)
    }
  }, [collaboratorId, collaboratorType, onDelete, onDeleteFailure, onDeleteSuccess])

  const initialValues = React.useMemo(() => {
    if (!collaborator) {
      return {
        collaborator_id: '',
        collaborator_type: 'user',
        rights: [...pseudoRights],
      }
    }

    return {
      collaborator_id: collaboratorId,
      collaborator_type: collaboratorType,
      rights: [...collaborator.rights],
    }
  }, [collaborator, collaboratorId, collaboratorType, pseudoRights])

  let warning = null
  if (update) {
    if (isCollaboratorCurrentUser) {
      warning = isCollaboratorAdmin ? (
        <Notification small warning content={sharedMessages.collaboratorWarningAdminSelf} />
      ) : (
        <Notification small warning content={sharedMessages.collaboratorWarningSelf} />
      )
    } else if (isCollaboratorAdmin) {
      warning = <Notification small warning content={sharedMessages.collaboratorWarningAdmin} />
    }
  }

  return (
    <Form
      error={error}
      onSubmit={handleSubmit}
      initialValues={initialValues}
      validationSchema={validationSchema}
    >
      {warning}
      <Form.Field
        name="collaborator_id"
        component={Input}
        title={sharedMessages.collaboratorId}
        placeholder={sharedMessages.collaboratorIdPlaceholder}
        required
        autoFocus={!update}
        disabled={update}
      />
      <Form.Field
        name="collaborator_type"
        title={sharedMessages.type}
        component={Radio.Group}
        disabled={update}
        required
      >
        <Radio label={sharedMessages.user} value="user" />
        <Radio label={sharedMessages.organization} value="organization" />
      </Form.Field>
      <Form.Field
        name="rights"
        title={sharedMessages.rights}
        required
        component={RightsGroup}
        rights={rights}
        pseudoRight={pseudoRights}
        entityTypeMessage={sharedMessages.collaborator}
      />
      <SubmitBar>
        <Form.Submit
          component={SubmitButton}
          message={update ? sharedMessages.saveChanges : sharedMessages.addCollaborator}
        />
        {update && (
          <ModalButton
            type="button"
            icon="delete"
            disabled={deleteDisabled}
            danger
            naked
            message={
              deleteDisabled
                ? sharedMessages.removeCollaboratorLast
                : isCollaboratorCurrentUser
                ? sharedMessages.removeCollaboratorSelf
                : sharedMessages.removeCollaborator
            }
            modalData={{
              message: isCollaboratorCurrentUser
                ? sharedMessages.collaboratorModalWarningSelf
                : {
                    values: { collaboratorId },
                    ...sharedMessages.collaboratorModalWarning,
                  },
            }}
            onApprove={handleDelete}
          />
        )}
      </SubmitBar>
    </Form>
  )
}

CollaboratorForm.propTypes = {
  collaborator: PropTypes.collaborator,
  collaboratorId: PropTypes.string,
  deleteDisabled: PropTypes.bool,
  error: PropTypes.error,
  isCollaboratorAdmin: PropTypes.bool.isRequired,
  isCollaboratorCurrentUser: PropTypes.bool.isRequired,
  isCollaboratorUser: PropTypes.bool.isRequired,
  onDelete: PropTypes.func,
  onDeleteFailure: PropTypes.func,
  onDeleteSuccess: PropTypes.func,
  onSubmit: PropTypes.func.isRequired,
  onSubmitFailure: PropTypes.func,
  onSubmitSuccess: PropTypes.func.isRequired,
  pseudoRights: PropTypes.rights,
  rights: PropTypes.rights.isRequired,
  update: PropTypes.bool,
}

CollaboratorForm.defaultProps = {
  deleteDisabled: false,
  onSubmitFailure: () => null,
  onDelete: () => null,
  onDeleteSuccess: () => null,
  onDeleteFailure: () => null,
  pseudoRights: [],
  error: undefined,
  collaborator: undefined,
  update: false,
  collaboratorId: undefined,
}

export default CollaboratorForm
