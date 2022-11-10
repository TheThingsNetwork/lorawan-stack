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

import React, { useCallback, useState, useRef, useEffect } from 'react'
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
  tokenDescription: 'Geolocation access token as configured within LoRa Cloud',
  setLoRaCloudToken: 'Set LoRa Cloud token',
  deleteWarning:
    'Are you sure you want to delete the LoRaCloud Geolocation token? This action cannot be undone.',
  queryType: 'Query type',
  queryTypeDescription: 'What kind of geolocation query should be used',
  multiFrameDescription: 'Enable multiframe lookups to improve accuracy',
  multiFrameWindowSize: 'Multiframe window size',
  multiFrameWindowSizeDescription: 'How many historical messages to send as part of the request.',
  multiFrameTimeWindow: 'Multiframe time window',
  multiFrameTimeWindowDescription: 'The maximum age of considered historical messages in minutes',
  determineWindowSizeAutomatically: 'Determine window size automatically',
  enableMultiFrame: 'Enable multiframe',
  automaticMultiFrameDescription:
    'Determines the count of sent historical messages considered for geolocation based on the first byte of the payload',
})

const LORACLOUD_GLS_QUERY_LABELS = Object.freeze([
  { value: 'TOARSSI', label: 'LoRa® TOA/RSSI' },
  { value: 'GNSS', label: 'GNSS' },
  { value: 'TOAWIFI', label: 'TOA/WiFi' },
])
const LORACLOUD_GLS_QUERY_TYPES = Object.freeze({
  TOARSSI: 'TOARSSI',
  GNSS: 'GNSS',
  TOAWIFI: 'TOAWIFI',
})
const LORACLOUD_GLS_QUERY_VALUES = Object.freeze(Object.values(LORACLOUD_GLS_QUERY_TYPES))

const validationSchema = Yup.object()
  .shape({
    data: Yup.object().shape({
      token: Yup.string().default('').required(sharedMessages.validateRequired),
      query: Yup.string()
        .oneOf(LORACLOUD_GLS_QUERY_VALUES)
        .default(LORACLOUD_GLS_QUERY_TYPES.TOARSSI)
        .required(sharedMessages.validateRequired),
      server_url: Yup.string().url(sharedMessages.validateUrl),
      multi_frame: Yup.boolean().when('query', {
        is: LORACLOUD_GLS_QUERY_TYPES.TOARSSI,
        then: schema => schema.default(false).required(sharedMessages.validateRequired),
        otherwise: schema => schema.strip(),
      }),
      multi_frame_window_size: Yup.number().when('multi_frame', {
        is: true,
        then: schema =>
          schema
            .min(0, Yup.passValues(sharedMessages.validateNumberGte))
            .max(16, Yup.passValues(sharedMessages.validateNumberLte))
            .default(0)
            .required(sharedMessages.validateRequired),
        otherwise: schema => schema.strip(),
      }),
      multi_frame_window_age: Yup.number().when('multi_frame', {
        is: true,
        then: schema =>
          schema
            .min(1, Yup.passValues(sharedMessages.validateNumberGte))
            .max(7 * 24 * 60, Yup.passValues(sharedMessages.validateNumberLte))
            .default(24 * 60)
            .required(sharedMessages.validateRequired),
        otherwise: schema => schema.strip(),
      }),
    }),
  })
  .noUnknown()

const promisifiedSetAppPkgDefaultAssoc = attachPromise(setAppPkgDefaultAssoc)
const promisifiedDeleteAppPkgDefaultAssoc = attachPromise(deleteAppPkgDefaultAssoc)

const decodeDetermineMultiframeAutomatically = value => value === 0
const encodeDetermineMultiframeAutomatically = value => (value ? 0 : 1)

const defaultValues = {
  data: {
    server_url: LORA_CLOUD_GLS.DEFAULT_SERVER_URL,
  },
}

const LoRaCloudGLSForm = () => {
  const [error, setError] = useState('')
  const appId = useSelector(selectSelectedApplicationId)
  const formRef = useRef(null)

  const dispatch = useDispatch()
  const defaultAssociation = useSelector(state =>
    selectApplicationPackageDefaultAssociation(state, LORA_CLOUD_GLS.DEFAULT_PORT),
  )
  const packageError = useSelector(selectGetApplicationPackagesError)
  const initialValues = validationSchema.cast(
    defaultAssociation ? { server_url: '', ...defaultAssociation } : defaultValues,
  )
  const handleSubmit = useCallback(
    async values => {
      try {
        const castedValues = validationSchema.cast(values)
        await dispatch(
          promisifiedSetAppPkgDefaultAssoc(appId, LORA_CLOUD_GLS.DEFAULT_PORT, {
            package_name: LORA_CLOUD_GLS.DEFAULT_PACKAGE_NAME,
            ...castedValues,
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
      formRef.current.resetForm({ values: validationSchema.getDefault() })
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

  const [queryType, setQueryType] = useState()
  const handleQueryTypeChange = useCallback(
    value => {
      setQueryType(value)
      const { setValues, values } = formRef.current
      setValues(validationSchema.cast(values))
    },
    [setQueryType, formRef],
  )

  const [multiFrame, setMultiFrame] = useState()
  const handleMultiFrameChange = useCallback(
    evt => {
      setMultiFrame(evt.target.checked)
      const { setValues, values } = formRef.current
      setValues(validationSchema.cast(values))
    },
    [setMultiFrame, formRef],
  )

  const [automaticMultiFrame, setAutomaticMultiFrame] = useState(
    initialValues.data.multi_frame_window_size === 0 ||
      initialValues.data.multi_frame_window_size === undefined,
  )
  const handleAutomaticWindowSizeChange = useCallback(
    evt => {
      const checked = evt.target.checked
      setAutomaticMultiFrame(checked)
    },
    [setAutomaticMultiFrame],
  )

  useEffect(() => {
    setQueryType(initialValues.data.query)
    setMultiFrame(initialValues.data.multi_frame)
  }, [initialValues.data.query, initialValues.data.multi_frame])

  return (
    <Form
      error={error}
      validationSchema={validationSchema}
      initialValues={initialValues}
      onSubmit={handleSubmit}
      formikRef={formRef}
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
        component={Select}
        title={m.queryType}
        description={m.queryTypeDescription}
        name="data.query"
        options={LORACLOUD_GLS_QUERY_LABELS}
        disabled={LORACLOUD_GLS_QUERY_LABELS.length === 1}
        onChange={handleQueryTypeChange}
        required
      />
      {queryType === LORACLOUD_GLS_QUERY_TYPES.TOARSSI && (
        <>
          <Form.Field
            component={Checkbox}
            label={m.enableMultiFrame}
            description={m.multiFrameDescription}
            name="data.multi_frame"
            onChange={handleMultiFrameChange}
          />
          {multiFrame && (
            <>
              <Form.Field
                component={Checkbox}
                label={m.determineWindowSizeAutomatically}
                description={m.automaticMultiFrameDescription}
                onChange={handleAutomaticWindowSizeChange}
                decode={decodeDetermineMultiframeAutomatically}
                encode={encodeDetermineMultiframeAutomatically}
                name="data.multi_frame_window_size"
              />
              {!automaticMultiFrame && (
                <Form.Field
                  component={Input}
                  title={m.multiFrameWindowSize}
                  description={m.multiFrameWindowSizeDescription}
                  name="data.multi_frame_window_size"
                  type="number"
                  min={1}
                  max={16}
                  inputWidth="xs"
                  required
                />
              )}
              <Form.Field
                component={Input}
                title={m.multiFrameTimeWindow}
                description={m.multiFrameTimeWindowDescription}
                name="data.multi_frame_window_age"
                type="number"
                min={1}
                max={7 * 24 * 60}
                inputWidth="xs"
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
  )
}

export default LoRaCloudGLSForm
