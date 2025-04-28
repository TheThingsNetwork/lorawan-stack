// Copyright © 2025 The Things Network Foundation, The Things Industries B.V.
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
import { defineMessages } from 'react-intl'

import tts from '@console/api/tts'
import { APPLICATION, GATEWAY, ORGANIZATION, USER } from '@console/constants/entities'

import FormField from '@ttn-lw/components/form/field'
import FormSubmit from '@ttn-lw/components/form/submit'
import SubmitButton from '@ttn-lw/components/submit-button'
import Input from '@ttn-lw/components/input'
import Button from '@ttn-lw/components/button'
import { useFormContext } from '@ttn-lw/components/form'
import RightsGroupModal from '@ttn-lw/components/rights-group-modal'
import toast from '@ttn-lw/components/toast'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import ApiKeyForm from './form'
import validationSchema from './validation-schema'
import { encodeExpiryDate, decodeExpiryDate } from './utils'

const sdkServices = {
  [APPLICATION]: 'Applications',
  [GATEWAY]: 'Gateways',
  [ORGANIZATION]: 'Organizations',
  [USER]: 'Users',
}

const m = defineMessages({
  createApiKeyError: 'Failed to create API key',
})

const FormButtons = ({ handleCancel }) => {
  const { isValid } = useFormContext()
  return (
    <div className="d-flex gap-cs-xs j-end mt-cs-xl">
      <Button message={sharedMessages.cancel} secondary onClick={handleCancel} />
      <FormSubmit
        component={SubmitButton}
        message={sharedMessages.createApiKey}
        disabled={!isValid}
      />
    </div>
  )
}

FormButtons.propTypes = {
  handleCancel: PropTypes.func.isRequired,
}

const CreateForm = ({ entity, entityId, handleCancel, rights, setApiKey }) => {
  const handleCreate = useCallback(
    async values => {
      const castedValues = validationSchema.cast(values)

      return await tts[sdkServices[entity]].ApiKeys.create(entityId, castedValues)
    },
    [entity, entityId],
  )

  const handleCreateSuccess = useCallback(
    key => {
      setApiKey(key.key)
    },
    [setApiKey],
  )

  const handleCreateFailure = useCallback(() => {
    toast({
      type: toast.types.ERROR,
      message: m.createApiKeyError,
    })
  }, [])

  const initialValues = {
    name: '',
    rights: [],
    expires_at: null,
  }

  return (
    <>
      <ApiKeyForm
        onSubmit={handleCreate}
        onSubmitSuccess={handleCreateSuccess}
        onSubmitFailure={handleCreateFailure}
        validationSchema={validationSchema}
        initialValues={initialValues}
      >
        <FormField
          title={sharedMessages.name}
          placeholder={sharedMessages.apiKeyNamePlaceholder}
          name="name"
          autoFocus
          component={Input}
          fieldWidth="full"
        />
        <FormField
          title={'Expiry date'}
          name="expires_at"
          type="date"
          encode={encodeExpiryDate}
          decode={decodeExpiryDate}
          component={Input}
          fieldWidth="full"
        />
        <FormField
          name="rights"
          title={sharedMessages.rights}
          required
          component={RightsGroupModal}
          rights={rights}
          fieldWidth="full"
        />
        <FormButtons handleCancel={handleCancel} />
      </ApiKeyForm>
    </>
  )
}

CreateForm.propTypes = {
  entity: PropTypes.entity.isRequired,
  entityId: PropTypes.string.isRequired,
  handleCancel: PropTypes.func.isRequired,
  rights: PropTypes.rightsRadioItems.isRequired,
  setApiKey: PropTypes.func.isRequired,
}

export default CreateForm
