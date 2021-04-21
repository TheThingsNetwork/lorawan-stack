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

import LORA_CLOUD_GLS from '@console/constants/lora-cloud-gls'

import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import SubmitBar from '@ttn-lw/components/submit-bar'
import SubmitButton from '@ttn-lw/components/submit-button'
import ModalButton from '@ttn-lw/components/button/modal-button'
import toast from '@ttn-lw/components/toast'
import Checkbox from '@ttn-lw/components/checkbox'
import Select from '@ttn-lw/components/select'

import RequireRequest from '@ttn-lw/lib/components/require-request'

import Yup from '@ttn-lw/lib/yup'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import { isNotFoundError } from '@ttn-lw/lib/errors/utils'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import {
  setAppPkgDefaultAssoc,
  getAppPkgDefaultAssoc,
  deleteAppPkgDefaultAssoc,
} from '@console/store/actions/application-packages'

import { selectSelectedApplicationId } from '@console/store/selectors/applications'
import {
  selectApplicationPackageDefaultAssociation,
  selectGetApplicationPackagesError,
} from '@console/store/selectors/application-packages'

const m = defineMessages({
  tokenDescription: 'Geolocation access token as configured within LoRa Cloud',
  setLoRaCloudToken: 'Set LoRa Cloud token',
  deleteWarning:
    'Are you sure you want to delete the LoRaCloud Geolocation token? This action cannot be undone.',
  queryType: 'Query type',
  queryTypeDescription: 'What kind of geolocation query should be used',
  multiFrame: 'Multiframe',
  multiFrameDescription: 'Enable multiframe lookups',
  multiFrameWindowSize: 'Multiframe window size',
  multiFrameWindowSizeDescription:
    'How many historical message to send as part of the request. A window size of 0 automatically determines this based on the first byte of the payload.',
  multiFrameTimeWindow: 'Multiframe time window',
  multiFrameTimeWindowDescription: 'How recent the historical messages should be, in minutes.',
})

const LORACLOUD_GLS_QUERY_LABELS = Object.freeze([{ value: 'TDOARSSI', label: 'LoRa® TOA/RSSI' }])
const LORACLOUD_GLS_QUERY_TYPES = Object.freeze({
  TDOARSSI: 'TDOARSSI',
})
const LORACLOUD_GLS_QUERY_VALUES = Object.freeze(Object.values(LORACLOUD_GLS_QUERY_TYPES))

const validationSchema = Yup.object()
  .shape({
    data: Yup.object().shape({
      token: Yup.string().required(sharedMessages.validateRequired),
      query: Yup.string()
        .oneOf(LORACLOUD_GLS_QUERY_VALUES)
        .required(sharedMessages.validateRequired),
      multi_frame: Yup.boolean().required(sharedMessages.validateRequired),
      multi_frame_window_size: Yup.number()
        .min(0, Yup.passValues(sharedMessages.validateNumberGte))
        .max(16, Yup.passValues(sharedMessages.validateNumberLte))
        .required(sharedMessages.validateRequired),
      multi_frame_window_age: Yup.number()
        .min(1, Yup.passValues(sharedMessages.validateNumberGte))
        .max(7 * 24 * 60, Yup.passValues(sharedMessages.validateNumberLte))
        .required(sharedMessages.validateRequired),
    }),
  })
  .noUnknown()

const defaultValues = {
  data: {
    token: '',
    query: LORACLOUD_GLS_QUERY_TYPES.TDOARSSI,
    multi_frame: false,
    multi_frame_window_size: 0,
    multi_frame_window_age: 1440,
  },
}

const promisifiedSetAppPkgDefaultAssoc = attachPromise(setAppPkgDefaultAssoc)
const promisifiedDeleteAppPkgDefaultAssoc = attachPromise(deleteAppPkgDefaultAssoc)

const LoRaCloudGLSForm = () => {
  const [error, setError] = useState('')
  const appId = useSelector(selectSelectedApplicationId)
  const selector = ['data']

  const dispatch = useDispatch()
  const defaultAssociation = useSelector(state =>
    selectApplicationPackageDefaultAssociation(state, LORA_CLOUD_GLS.DEFAULT_PORT),
  )
  const packageError = useSelector(selectGetApplicationPackagesError)
  const initialValues = validationSchema.cast(defaultAssociation || defaultValues)

  const handleSubmit = useCallback(
    async values => {
      try {
        await dispatch(
          promisifiedSetAppPkgDefaultAssoc(appId, LORA_CLOUD_GLS.DEFAULT_PORT, {
            package_name: LORA_CLOUD_GLS.DEFAULT_PACKAGE_NAME,
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
        promisifiedDeleteAppPkgDefaultAssoc(appId, LORA_CLOUD_GLS.DEFAULT_PORT, {
          package_name: LORA_CLOUD_GLS.DEFAULT_PACKAGE_NAME,
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

  const initialQuery = initialValues.data.query
  const [queryType, setQueryType] = useState(initialQuery)

  const initialMultiFrame = initialValues.data.multi_frame
  const [multiFrame, setMultiFrame] = useState(initialMultiFrame)
  const handleMultiFrameChange = useCallback(evt => setMultiFrame(evt.target.checked), [
    setMultiFrame,
  ])

  return (
    <RequireRequest
      requestAction={getAppPkgDefaultAssoc(appId, LORA_CLOUD_GLS.DEFAULT_PORT, selector)}
    >
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
        <Form.Field
          component={Select}
          title={m.queryType}
          description={m.queryTypeDescription}
          name="data.query"
          options={LORACLOUD_GLS_QUERY_LABELS}
          onChange={setQueryType}
          required
        />
        {queryType === LORACLOUD_GLS_QUERY_TYPES.TDOARSSI && (
          <>
            <Form.Field
              component={Checkbox}
              title={m.multiFrame}
              description={m.multiFrameDescription}
              name="data.multi_frame"
              onChange={handleMultiFrameChange}
              required
            />
            {multiFrame && (
              <>
                <Form.Field
                  component={Input}
                  title={m.multiFrameWindowSize}
                  description={m.multiFrameWindowSizeDescription}
                  name="data.multi_frame_window_size"
                  type="number"
                  min={0}
                  max={16}
                  required
                />
                <Form.Field
                  component={Input}
                  title={m.multiFrameTimeWindow}
                  description={m.multiFrameTimeWindowDescription}
                  name="data.multi_frame_window_age"
                  type="number"
                  min={1}
                  max={7 * 24 * 60}
                  required
                />
              </>
            )}
          </>
        )}
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
    </RequireRequest>
  )
}

export default LoRaCloudGLSForm
