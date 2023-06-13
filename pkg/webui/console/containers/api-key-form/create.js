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
import { useNavigate } from 'react-router-dom'

import SubmitBar from '@ttn-lw/components/submit-bar'
import FormField from '@ttn-lw/components/form/field'
import FormSubmit from '@ttn-lw/components/form/submit'
import SubmitButton from '@ttn-lw/components/submit-button'
import Input from '@ttn-lw/components/input'
import RightsGroup from '@ttn-lw/components/rights-group'

import ApiKeyModal from '@console/components/api-key-modal'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import ApiKeyForm from './form'
import validationSchema from './validation-schema'
import { encodeExpiryDate, decodeExpiryDate } from './utils'
import useApiKeyData from './hooks'

const CreateForm = ({ entity, entityId }) => {
  const [modal, setModal] = useState(null)
  const { rights, pseudoRights, create } = useApiKeyData(entity)
  const navigate = useNavigate()

  const handleModalApprove = useCallback(async () => {
    setModal(null)
    // Navigate back to list
    navigate('../')
  }, [navigate])

  const handleCreate = useCallback(
    async values => {
      const castedValues = validationSchema.cast(values)

      return await create(entityId, castedValues)
    },
    [create, entityId],
  )

  const handleCreateSuccess = useCallback(
    key => {
      setModal({
        secret: key.key,
        rights: key.rights,
        onComplete: handleModalApprove,
        approval: false,
      })
    },
    [handleModalApprove],
  )

  const handleCreateFailure = useCallback(() => null, [])

  const initialValues = {
    name: '',
    rights: [...pseudoRights],
    expires_at: null,
  }

  const modalProps = modal || {}
  const modalVisible = Boolean(modal)

  return (
    <>
      <ApiKeyModal {...modalProps} visible={modalVisible} approval={false} />
      <ApiKeyForm
        rights={rights}
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
        />
        <FormField
          title={'Expiry date'}
          name="expires_at"
          type="date"
          encode={encodeExpiryDate}
          decode={decodeExpiryDate}
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
          <FormSubmit component={SubmitButton} message={sharedMessages.createApiKey} />
        </SubmitBar>
      </ApiKeyForm>
    </>
  )
}

CreateForm.propTypes = {
  entity: PropTypes.entity.isRequired,
  entityId: PropTypes.string.isRequired,
}

export default CreateForm
