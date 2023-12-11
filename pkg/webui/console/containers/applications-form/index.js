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
import { defineMessages } from 'react-intl'
import { useDispatch, useSelector } from 'react-redux'

import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import SubmitButton from '@ttn-lw/components/submit-button'
import SubmitBar from '@ttn-lw/components/submit-bar'

import OwnersSelect from '@console/containers/owners-select'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import Yup from '@ttn-lw/lib/yup'
import { id as applicationIdRegexp } from '@ttn-lw/lib/regexp'
import { selectAsConfig, selectJsConfig, selectNsConfig } from '@ttn-lw/lib/selectors/env'
import { getApplicationId } from '@ttn-lw/lib/selectors/id'
import getHostFromUrl from '@ttn-lw/lib/host-from-url'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import PropTypes from '@ttn-lw/lib/prop-types'

import { createApp } from '@console/store/actions/applications'

import { selectUserId } from '@account/store/selectors/user'

const m = defineMessages({
  applicationName: 'Application name',
  appIdPlaceholder: 'my-new-application',
  appNamePlaceholder: 'My new application',
  appDescPlaceholder: 'Description for my new application',
  appDescDescription:
    'Optional application description; can also be used to save notes about the application',
  appDescription:
    'Within applications, you can register and manage end devices, aggregate their sensor data and act on it using our many integration options.{break}Learn more in our <Link>Applications Guide</Link>',
})

const validationSchema = Yup.object().shape({
  owner_id: Yup.string().required(sharedMessages.validateRequired),
  application_id: Yup.string()
    .min(3, Yup.passValues(sharedMessages.validateTooShort))
    .max(36, Yup.passValues(sharedMessages.validateTooLong))
    .matches(applicationIdRegexp, Yup.passValues(sharedMessages.validateIdFormat))
    .required(sharedMessages.validateRequired),
  name: Yup.string()
    .min(2, Yup.passValues(sharedMessages.validateTooShort))
    .max(2000, Yup.passValues(sharedMessages.validateTooLong)),
})

const ApplicationForm = props => {
  const { onSuccess } = props
  const jsHost = getHostFromUrl(selectJsConfig().base_url)
  const nsHost = getHostFromUrl(selectNsConfig().base_url)
  const asHost = getHostFromUrl(selectAsConfig().base_url)

  const [error, setError] = useState()

  const dispatch = useDispatch()
  const userId = useSelector(selectUserId)

  const initialValues = {
    application_id: '',
    name: '',
    description: '',
    owner_id: userId,
  }
  const handleSubmit = useCallback(
    async values => {
      const { owner_id, application_id, name, description } = values

      setError(undefined)

      try {
        const result = await dispatch(
          attachPromise(
            createApp(
              owner_id,
              {
                ids: { application_id },
                name,
                description,
                network_server_address: nsHost,
                application_server_address: asHost,
                join_server_address: jsHost,
              },
              userId === owner_id,
            ),
          ),
        )

        const appId = getApplicationId(result)
        onSuccess(appId)
      } catch (error) {
        setError(error)
      }
    },
    [dispatch, nsHost, asHost, jsHost, userId, onSuccess],
  )

  return (
    <Form
      error={error}
      onSubmit={handleSubmit}
      initialValues={initialValues}
      validationSchema={validationSchema}
    >
      <OwnersSelect name="owner_id" required />
      <Form.Field
        title={sharedMessages.appId}
        name="application_id"
        placeholder={m.appIdPlaceholder}
        required
        component={Input}
        autoFocus
      />
      <Form.Field
        title={m.applicationName}
        name="name"
        placeholder={m.appNamePlaceholder}
        component={Input}
      />
      <Form.Field
        title={sharedMessages.description}
        type="textarea"
        name="description"
        placeholder={m.appDescPlaceholder}
        description={m.appDescDescription}
        component={Input}
      />
      <SubmitBar>
        <Form.Submit message={sharedMessages.createApplication} component={SubmitButton} />
      </SubmitBar>
    </Form>
  )
}

ApplicationForm.propTypes = {
  onSuccess: PropTypes.func.isRequired,
}

export default ApplicationForm
