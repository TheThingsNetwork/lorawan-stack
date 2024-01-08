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

import React, { useCallback, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { defineMessages } from 'react-intl'
import { isEmpty } from 'lodash'

import Form from '@ttn-lw/components/form'
import Notification from '@ttn-lw/components/notification'
import SubmitBar from '@ttn-lw/components/submit-bar'
import SubmitButton from '@ttn-lw/components/submit-button'
import toast from '@ttn-lw/components/toast'
import ModalButton from '@ttn-lw/components/button/modal-button'
import RightsGroup from '@ttn-lw/components/rights-group'

import { composeOption } from '@ttn-lw/containers/collaborator-select/util'

import AccountSelect from '@console/containers/account-select'

import Yup from '@ttn-lw/lib/yup'
import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import { userId as collaboratorIdRegexp } from '@ttn-lw/lib/regexp'

import useCollaboratorData from './hooks'

const emptyCollaboratorCheck = collab =>
  !(collab === '') &&
  !(collab === undefined) &&
  !(collab === null) &&
  !(collab instanceof Object && Object.values(collab).every(val => !Boolean(val) || isEmpty(val)))

const collaboratorOrganizationSchema = Yup.object().shape({
  organization_id: Yup.string().matches(collaboratorIdRegexp, sharedMessages.validateAlphanum),
})

const collaboratorUserSchema = Yup.object().shape({
  user_id: Yup.string().matches(collaboratorIdRegexp, sharedMessages.validateAlphanum),
})

const validationSchema = Yup.object().shape({
  collaborator: Yup.object()
    .shape({
      ids: Yup.object().when(['organization_ids'], {
        is: organizationIds => Boolean(organizationIds),
        then: schema => schema.concat(collaboratorOrganizationSchema),
        otherwise: schema => schema.concat(collaboratorUserSchema),
      }),
    })
    .test('collaborator is not empty', sharedMessages.validateRequired, emptyCollaboratorCheck),
  rights: Yup.array().min(1, sharedMessages.validateRights),
})

const m = defineMessages({
  collaboratorIdPlaceholder: 'Type to choose a collaborator',
})

const encodeCollaborator = collaboratorOption =>
  collaboratorOption
    ? {
        ids: {
          [`${collaboratorOption.icon}_ids`]: {
            [`${collaboratorOption.icon}_id`]: collaboratorOption.value,
          },
        },
      }
    : null

const decodeCollaborator = collaborator =>
  collaborator && collaborator.ids ? composeOption(collaborator) : null

const CollaboratorForm = props => {
  const { entity, entityId, collaboratorId, deleteDisabled, update, tts } = props

  const {
    collaborator,
    isCollaboratorUser,
    isCollaboratorAdmin,
    isCollaboratorCurrentUser,
    error: passedError,
    rights,
    pseudoRights,
    updateCollaborator,
    removeCollaborator,
  } = useCollaboratorData(entity, entityId, collaboratorId, tts)

  const [submitError, setSubmitError] = useState(undefined)
  const navigate = useNavigate()
  const error = submitError || passedError

  const handleSubmit = useCallback(
    async (values, { resetForm, setSubmitting }) => {
      const { collaborator, rights } = values

      const composedCollaborator = {
        ...collaborator,
        rights,
      }

      setSubmitError(undefined)

      try {
        await updateCollaborator(composedCollaborator)

        resetForm({ values })
        if (!update) {
          navigate('..')
        } else {
          toast({
            message: sharedMessages.collaboratorUpdateSuccess,
            type: toast.types.SUCCESS,
          })
        }
      } catch (error) {
        setSubmitting(false)
        setSubmitError(error)
      }
    },
    [navigate, update, updateCollaborator],
  )
  const handleDelete = useCallback(async () => {
    setSubmitError(undefined)

    try {
      await removeCollaborator(isCollaboratorUser, collaboratorId)
      toast({
        message: sharedMessages.collaboratorDeleteSuccess,
        type: toast.types.SUCCESS,
      })
      navigate('../')
    } catch (error) {
      setSubmitError(error)
    }
  }, [collaboratorId, isCollaboratorUser, navigate, removeCollaborator])

  const initialValues = React.useMemo(() => {
    if (!collaborator) {
      return {
        collaborator: '',
        rights: [...pseudoRights],
      }
    }

    return {
      collaborator,
      rights: [...collaborator.rights],
    }
  }, [collaborator, pseudoRights])

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
      <AccountSelect
        name="collaborator"
        title={sharedMessages.collaborator}
        placeholder={m.collaboratorIdPlaceholder}
        noOptionsMessage={sharedMessages.noMatchingCollaborators}
        required
        autoFocus={!update}
        disabled={update}
        entity={entity}
        encode={encodeCollaborator}
        decode={decodeCollaborator}
      />
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
  collaboratorId: PropTypes.string,
  deleteDisabled: PropTypes.bool,
  entity: PropTypes.entity.isRequired,
  entityId: PropTypes.string.isRequired,
  tts: PropTypes.object.isRequired,
  update: PropTypes.bool,
}

CollaboratorForm.defaultProps = {
  collaboratorId: undefined,
  deleteDisabled: false,
  update: false,
}

export default CollaboratorForm
