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

import React from 'react'
import { Container, Col, Row } from 'react-grid-system'
import { connect } from 'react-redux'
import bind from 'autobind-decorator'
import { defineMessages } from 'react-intl'
import { push } from 'connected-react-router'

import api from '@console/api'

import PageTitle from '@ttn-lw/components/page-title'
import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import SubmitButton from '@ttn-lw/components/submit-button'
import SubmitBar from '@ttn-lw/components/submit-bar'

import OwnersSelect from '@console/containers/owners-select'

import withFeatureRequirement from '@console/lib/components/with-feature-requirement'

import Yup from '@ttn-lw/lib/yup'
import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import { getApplicationId } from '@ttn-lw/lib/selectors/id'
import { id as applicationIdRegexp } from '@ttn-lw/lib/regexp'

import { mayCreateApplications } from '@console/lib/feature-checks'

import { selectUserId, selectUserRights } from '@console/store/selectors/user'

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

@withFeatureRequirement(mayCreateApplications, { redirect: '/applications' })
@connect(
  state => ({
    userId: selectUserId(state),
    rights: selectUserRights(state),
  }),
  dispatch => ({
    navigateToApplication: appId => dispatch(push(`/applications/${appId}`)),
  }),
)
export default class Add extends React.Component {
  static propTypes = {
    navigateToApplication: PropTypes.func.isRequired,
    userId: PropTypes.string.isRequired,
  }

  constructor(props) {
    super(props)
    this.state = {
      error: '',
    }
  }

  @bind
  async handleSubmit(values, { setSubmitting }) {
    const { userId, navigateToApplication } = this.props
    const { owner_id, application_id, name, description } = values

    await this.setState({ error: '' })

    try {
      const result = await api.application.create(
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

      await this.setState({ error })
    }
  }

  render() {
    const { error } = this.state
    const { userId } = this.props

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
              onSubmit={this.handleSubmit}
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
}
