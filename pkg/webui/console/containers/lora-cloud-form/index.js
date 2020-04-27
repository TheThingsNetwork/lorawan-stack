// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
import { defineMessages } from 'react-intl'
import { useDispatch, useSelector } from 'react-redux'

import LORA_CLOUD_DAS from '@console/constants/lora-cloud-das'

import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import SubmitBar from '@ttn-lw/components/submit-bar'
import SubmitButton from '@ttn-lw/components/submit-button'
import ModalButton from '@ttn-lw/components/button/modal-button'
import toast from '@ttn-lw/components/toast'

import RequireRequest from '@ttn-lw/lib/components/require-request'

import Yup from '@ttn-lw/lib/yup'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import {
  setAppPkgDefaultAssoc,
  getAppPkgDefaultAssoc,
} from '@console/store/actions/application-packages'

import { selectSelectedApplicationId } from '@console/store/selectors/applications'
import { selectApplicationPackageDefaultAssociation } from '@console/store/selectors/application-packages'

const m = defineMessages({
  token: 'Token',
  setLoRaCloudToken: 'Set LoRa Cloud token',
  deleteToken: 'Delete token',
  deleteWarning: 'Are you sure you want to delete the current token? This action cannot be undone.',
  tokenUpdated: 'Token updated',
  tokenDeleted: 'Token deleted',
  setToken: 'Set token',
})

const validationSchema = Yup.object()
  .shape({
    data: Yup.object().shape({
      token: Yup.string().required(sharedMessages.validateRequired),
    }),
  })
  .noUnknown()

const defaultValues = {
  data: {
    token: '',
  },
}

const LoRaCloudForm = () => {
  const [error, setError] = useState('')
  const appId = useSelector(selectSelectedApplicationId)
  const selector = ['data']

  const dispatch = useDispatch()
  const defaultAssociation = useSelector(selectApplicationPackageDefaultAssociation)
  const initialValues = validationSchema.cast(defaultAssociation || defaultValues)

  const handleSubmit = useCallback(
    async values => {
      try {
        const result = await dispatch(
          attachPromise(setAppPkgDefaultAssoc)(appId, LORA_CLOUD_DAS.DEFAULT_PORT, {
            package_name: LORA_CLOUD_DAS.DEFAULT_PACKAGE_NAME,
            ...values,
          }),
        )
        const deleted = Boolean(!result.data.token)
        toast({
          title: 'LoRa Cloud',
          message: deleted ? m.tokenDeleted : m.tokenUpdated,
          type: toast.types.SUCCESS,
        })
      } catch (error) {
        setError(error)
      }
    },
    [appId, dispatch],
  )

  const handleDelete = useCallback(async () => await handleSubmit({ data: { token: '' } }), [
    handleSubmit,
  ])

  return (
    <RequireRequest
      requestAction={getAppPkgDefaultAssoc(appId, LORA_CLOUD_DAS.DEFAULT_PORT, selector)}
    >
      <Form
        error={error}
        validationSchema={validationSchema}
        initialValues={initialValues}
        onSubmit={handleSubmit}
        enableReinitialize
      >
        <Form.Field component={Input} title={m.token} name="data.token" required />
        <SubmitBar>
          <Form.Submit component={SubmitButton} message={m.setToken} />
          {Boolean(defaultAssociation) && (
            <ModalButton
              type="button"
              icon="delete"
              message={m.deleteToken}
              modalData={{
                message: m.deleteWarning,
              }}
              onApprove={handleDelete}
              danger
              naked
            />
          )}
        </SubmitBar>
      </Form>
    </RequireRequest>
  )
}

export default LoRaCloudForm
