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
import Checkbox from '@ttn-lw/components/checkbox'
import SubmitBar from '@ttn-lw/components/submit-bar'
import SubmitButton from '@ttn-lw/components/submit-button'
import ModalButton from '@ttn-lw/components/button/modal-button'
import toast from '@ttn-lw/components/toast'

import Yup from '@ttn-lw/lib/yup'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import { isNotFoundError } from '@ttn-lw/lib/errors/utils'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import {
  setAppPkgDefaultAssoc,
  deleteAppPkgDefaultAssoc,
} from '@console/store/actions/application-packages'

import { selectSelectedApplicationId } from '@console/store/selectors/applications'
import {
  selectApplicationPackageDefaultAssociation,
  selectGetApplicationPackagesError,
} from '@console/store/selectors/application-packages'

const m = defineMessages({
  tokenDescription: 'Device & Application Services Access Token as configured within LoRa Cloud',
  lr1110Encoding: 'LR1110 demo encoding',
  setLoRaCloudToken: 'Set LoRa Cloud token',
  deleteWarning:
    'Are you sure you want to delete the LoRa Cloud Device & Application Services token? This action cannot be undone.',
})

const validationSchema = Yup.object()
  .shape({
    data: Yup.object().shape({
      token: Yup.string().required(sharedMessages.validateRequired),
      use_tlv_encoding: Yup.boolean(),
    }),
  })
  .noUnknown()

const defaultValues = {
  data: {
    token: '',
  },
}

const promisifiedSetAppPkgDefaultAssoc = attachPromise(setAppPkgDefaultAssoc)
const promisifiedDeleteAppPkgDefaultAssoc = attachPromise(deleteAppPkgDefaultAssoc)

const LoRaCloudDASForm = () => {
  const [error, setError] = useState('')
  const appId = useSelector(selectSelectedApplicationId)

  const dispatch = useDispatch()
  const defaultAssociation = useSelector(state =>
    selectApplicationPackageDefaultAssociation(state, LORA_CLOUD_DAS.DEFAULT_PORT),
  )
  const packageError = useSelector(selectGetApplicationPackagesError)
  const initialValues = validationSchema.cast(defaultAssociation || defaultValues)

  const handleSubmit = useCallback(
    async values => {
      try {
        await dispatch(
          promisifiedSetAppPkgDefaultAssoc(appId, LORA_CLOUD_DAS.DEFAULT_PORT, {
            package_name: LORA_CLOUD_DAS.DEFAULT_PACKAGE_NAME,
            ...values,
          }),
        )
        toast({
          title: 'LoRa Cloud',
          message: sharedMessages.tokenUpdated,
          type: toast.types.SUCCESS,
        })
      } catch (error) {
        setError(error)
      }
    },
    [appId, dispatch],
  )

  const handleDelete = useCallback(async () => {
    try {
      await dispatch(
        promisifiedDeleteAppPkgDefaultAssoc(appId, LORA_CLOUD_DAS.DEFAULT_PORT, {
          package_name: LORA_CLOUD_DAS.DEFAULT_PACKAGE_NAME,
        }),
      )
      toast({
        title: 'LoRa Cloud',
        message: sharedMessages.tokenDeleted,
        type: toast.types.SUCCESS,
      })
    } catch (error) {
      setError(error)
    }
  }, [appId, dispatch])

  if (packageError && !isNotFoundError(packageError)) {
    throw error
  }

  return (
    <Form
      error={error}
      validationSchema={validationSchema}
      initialValues={initialValues}
      onSubmit={handleSubmit}
      enableReinitialize
    >
      <Form.Field
        component={Input}
        title={sharedMessages.token}
        description={m.tokenDescription}
        name="data.token"
        required
      />
      <Form.Field component={Checkbox} title={m.lr1110Encoding} name="data.use_tlv_encoding" />
      <SubmitBar>
        <Form.Submit component={SubmitButton} message={sharedMessages.tokenSet} />
        {Boolean(defaultAssociation) && (
          <ModalButton
            type="button"
            icon="delete"
            message={sharedMessages.tokenDelete}
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
  )
}

export default LoRaCloudDASForm
