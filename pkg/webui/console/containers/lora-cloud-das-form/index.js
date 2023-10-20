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
import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'

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
  fPortSetTitle: 'FPort Set',
  fPortSetDescription:
    'Comma separated list of F-Port values (1-223) to be used for LoRa Cloud Modem Services',
  modemEncoding: 'LoRa Edge Reference Tracker (Modem-E) encoding',
  deleteWarning:
    'Are you sure you want to delete the LoRa Cloud Modem and Geolocation Services token? This action cannot be undone.',
  fPortSetValidationFormat:
    'The FPort must be a number between 1 and 223, or a comma-separated list of numbers between 1 and 223',
})

const mapFPortInputToNumberArr = value => {
  if (!value || value === '') {
    return []
  }
  return value.split(',').map(v => Number.parseInt(v.trim()))
}

const validationSchema = Yup.object()
  .shape({
    data: Yup.object().shape({
      token: Yup.string().required(sharedMessages.validateRequired),
      use_tlv_encoding: Yup.boolean(),
      server_url: Yup.string().url(sharedMessages.validateUrl),
      f_port_set: Yup.string()
        .transform(value => {
          let returning = value
          if (Array.isArray(value)) {
            returning = value.join(',')
          }
          return returning
        })
        .test('fport-format', m.fPortSetValidationFormat, value => {
          // Blank value or comma-separated list of numbers between 1 and 223
          const matchesFormat = value.match(/^$|^\d+(\s*,\s*\d+)*$/)
          if (!matchesFormat) {
            return false
          }
          const fPorts = mapFPortInputToNumberArr(value)
          return fPorts.every(fPort => fPort >= 1 && fPort <= 223)
        })
        .default(''),
    }),
  })
  .noUnknown()

const defaultValues = {
  data: {
    token: '',
    server_url: LORA_CLOUD_DAS.DEFAULT_SERVER_URL,
    f_port_set: LORA_CLOUD_DAS.DEFAULT_PORT_SET,
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
  const initialValues = validationSchema.cast(
    defaultAssociation ? { server_url: '', ...defaultAssociation } : defaultValues,
  )

  const handleSubmit = useCallback(
    async values => {
      try {
        values.data.f_port_set = mapFPortInputToNumberArr(values.data.f_port_set)
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
        sensitive
        required
      />
      <Form.Field
        component={Input}
        title={sharedMessages.serverUrl}
        description={sharedMessages.loraCloudServerUrlDescription}
        name="data.server_url"
      />
      <Form.Field
        component={Checkbox}
        title={m.modemEncoding}
        tooltipId={tooltipIds.LORA_CLOUD_MODEM_ENCODING}
        name="data.use_tlv_encoding"
      />
      <Form.Field
        component={Input}
        title={m.fPortSetTitle}
        description={m.fPortSetDescription}
        name="data.f_port_set"
      />
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
