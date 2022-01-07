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

import React, { useState, useCallback } from 'react'
import { Container, Col, Row } from 'react-grid-system'
import { defineMessages } from 'react-intl'

import tts from '@console/api/tts'

import PageTitle from '@ttn-lw/components/page-title'
import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import SubmitButton from '@ttn-lw/components/submit-button'
import SubmitBar from '@ttn-lw/components/submit-bar'

import OwnersSelect from '@console/containers/owners-select'

import Yup from '@ttn-lw/lib/yup'
import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import { getApplicationId } from '@ttn-lw/lib/selectors/id'
import { id as applicationIdRegexp } from '@ttn-lw/lib/regexp'

const m = defineMessages({
  applicationName: 'Application name',
  appIdPlaceholder: 'my-new-application',
  appNamePlaceholder: 'My new application',
  appDescPlaceholder: 'Description for my new application',
  appDescDescription:
    'Optional application description; can also be used to save notes about the application',
  createApplication: 'Create application',
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
  description: Yup.string(),
})

const ApplicationAdd = props => {
  const { userId, navigateToApplication } = props

  const [error, setError] = useState()

  const handleSubmit = useCallback(
    async (values, { setSubmitting }) => {
      const { owner_id, application_id, name, description } = values

      setError(undefined)

      try {
        const result = await tts.Applications.create(
          owner_id,
          {
            ids: { application_id },
            name,
            description,
          },
          userId === owner_id,
        )

        const appId = getApplicationId(result)

        navigateToApplication(appId)
      } catch (error) {
        setSubmitting(false)
        setError(error)
      }
    },
    [navigateToApplication, userId],
  )

  const initialValues = {
    application_id: '',
    name: '',
    description: '',
    owner_id: userId,
  }

  return (
    <Container>
      <PageTitle tall title={sharedMessages.addApplication} />
      <Row>
        <Col md={10} lg={9}>
          <Form
            error={error}
            onSubmit={handleSubmit}
            initialValues={initialValues}
            validationSchema={validationSchema}
          >
            <OwnersSelect name="owner_id" required autoFocus />
            <Form.Field
              title={sharedMessages.appId}
              name="application_id"
              placeholder={m.appIdPlaceholder}
              required
              component={Input}
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
              <Form.Submit message={m.createApplication} component={SubmitButton} />
            </SubmitBar>
          </Form>
        </Col>
      </Row>
    </Container>
  )
}

ApplicationAdd.propTypes = {
  navigateToApplication: PropTypes.func.isRequired,
  userId: PropTypes.string.isRequired,
}

export default ApplicationAdd
