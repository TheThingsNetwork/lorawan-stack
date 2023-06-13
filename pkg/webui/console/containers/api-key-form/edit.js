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

import React, { useState, useCallback } from 'react'
import { defineMessages } from 'react-intl'
import { useNavigate } from 'react-router-dom'

import SubmitBar from '@ttn-lw/components/submit-bar'
import ModalButton from '@ttn-lw/components/button/modal-button'
import toast from '@ttn-lw/components/toast'
import FormField from '@ttn-lw/components/form/field'
import FormSubmit from '@ttn-lw/components/form/submit'
import SubmitButton from '@ttn-lw/components/submit-button'
import Input from '@ttn-lw/components/input'
import RightsGroup from '@ttn-lw/components/rights-group'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import ApiKeyForm from './form'
import validationSchema from './validation-schema'
import { encodeExpiryDate, decodeExpiryDate } from './utils'
import useApiKeyData from './hooks'

const m = defineMessages({
  deleteKey: 'Delete key',
  modalWarning:
    'Are you sure you want to delete the {keyName} API key? Deleting an API key cannot be undone.',
  updateSuccess: 'API key updated',
  deleteSuccess: 'API key deleted',
})

const ApiKeyEditForm = ({ entity, entityId }) => {
  const [error, setError] = useState(null)
  const navigate = useNavigate()
  const { rights, pseudoRights, updateById, deleteById, apiKey } = useApiKeyData(entity, entityId)

  const handleEdit = useCallback(
    async values => {
      const castedValues = validationSchema.cast(values)

      return await updateById(castedValues)
    },
    [updateById],
  )

  const handleDeleteSuccess = useCallback(async () => {
    toast({
      message: m.deleteSuccess,
      type: toast.types.SUCCESS,
    })
    navigate('../')
  }, [navigate])

  const handleEditSuccess = useCallback(async () => {
    toast({
      message: m.updateSuccess,
      type: toast.types.SUCCESS,
    })
  }, [])

  const handleDelete = useCallback(async () => {
    setError(null)

    try {
      await deleteById(apiKey.id)
      await handleDeleteSuccess(apiKey.id)
    } catch (error) {
      setError(error)
    }
  }, [deleteById, apiKey.id, handleDeleteSuccess])

  const initialValues = {
    id: apiKey.id,
    name: apiKey.name,
    rights: apiKey.rights,
    expires_at: apiKey.expires_at,
  }

  return (
    <ApiKeyForm
      formError={error}
      initialValues={initialValues}
      validationSchema={validationSchema}
      onSubmit={handleEdit}
      onSubmitSuccess={handleEditSuccess}
      onSubmitFailure={setError}
    >
      <FormField title={sharedMessages.keyId} name="id" required valid disabled component={Input} />
      <FormField
        title={sharedMessages.name}
        placeholder={sharedMessages.apiKeyNamePlaceholder}
        name="name"
        component={Input}
      />
      <FormField
        title={'Expiry date'}
        name="expires_at"
        type="date"
        decode={decodeExpiryDate}
        encode={encodeExpiryDate}
        component={Input}
      />
      <FormField
        name="rights"
        title={sharedMessages.rights}
        required
        component={RightsGroup}
        rights={rights}
        pseudoRight={pseudoRights}
        entityTypeMessage={sharedMessages.apiKey}
      />
      <SubmitBar>
        <FormSubmit component={SubmitButton} message={sharedMessages.saveChanges} />
        <ModalButton
          type="button"
          icon="delete"
          danger
          naked
          message={m.deleteKey}
          modalData={{
            message: {
              values: { keyName: apiKey.name ? `"${apiKey.name}"` : '' },
              ...m.modalWarning,
            },
          }}
          onApprove={handleDelete}
        />
      </SubmitBar>
    </ApiKeyForm>
  )
}

ApiKeyEditForm.propTypes = {
  entity: PropTypes.entity.isRequired,
  entityId: PropTypes.string.isRequired,
}

export default ApiKeyEditForm
