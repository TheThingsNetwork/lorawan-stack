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
import bind from 'autobind-decorator'
import { defineMessages } from 'react-intl'
import * as Yup from 'yup'

import Form from '../../../components/form'
import Input from '../../../components/input'
import SubmitBar from '../../../components/submit-bar'
import SubmitButton from '../../../components/submit-button'

import { id as organizationIdRegexp } from '../../lib/regexp'
import PropTypes from '../../../lib/prop-types'
import sharedMessages from '../../../lib/shared-messages'

const validationSchema = Yup.object().shape({
  ids: Yup.object().shape({
    organization_id: Yup.string()
      .matches(organizationIdRegexp, sharedMessages.validateAlphanum)
      .min(2, sharedMessages.validateTooShort)
      .max(25, sharedMessages.validateTooLong)
      .required(sharedMessages.validateRequired),
  }),
  name: Yup.string()
    .min(2, sharedMessages.validateTooShort)
    .max(50, sharedMessages.validateTooLong),
  description: Yup.string().max(2000, sharedMessages.validateTooLong),
})

const m = defineMessages({
  createOrganization: 'Create Organization',
  orgDescPlaceholder: 'Description for my new organization',
  orgIdPlaceholder: 'my-new-organization',
  orgNamePlaceholder: 'My New Organization',
})

class OrganizationForm extends React.Component {
  static propTypes = {
    error: PropTypes.error,
    initialValues: PropTypes.shape({
      ids: PropTypes.shape({
        organization_id: PropTypes.string.isRequired,
      }).isRequired,
      name: PropTypes.string,
      description: PropTypes.string,
    }).isRequired,
    onSubmit: PropTypes.func.isRequired,
    onSubmitFailure: PropTypes.func,
    onSubmitSuccess: PropTypes.func,
  }

  static defaultProps = {
    error: '',
    onSubmitFailure: () => null,
    onSubmitSuccess: () => null,
  }

  @bind
  async handleSubmit(values, { resetForm }) {
    const { onSubmit, onSubmitSuccess, onSubmitFailure } = this.props
    const castedValues = validationSchema.cast(values)

    try {
      const result = await onSubmit(castedValues)
      onSubmitSuccess(result)
    } catch (error) {
      resetForm(values)
      onSubmitFailure(error)
    }
  }

  render() {
    const { error, initialValues } = this.props

    return (
      <Form
        error={error}
        onSubmit={this.handleSubmit}
        initialValues={initialValues}
        validationSchema={validationSchema}
      >
        <Form.Field
          title={sharedMessages.organizationId}
          name="ids.organization_id"
          placeholder={m.orgIdPlaceholder}
          autoFocus
          required
          component={Input}
        />
        <Form.Field
          title={sharedMessages.name}
          name="name"
          placeholder={m.orgNamePlaceholder}
          component={Input}
        />
        <Form.Field
          title={sharedMessages.description}
          name="description"
          type="textarea"
          placeholder={m.orgDescPlaceholder}
          component={Input}
        />
        <SubmitBar>
          <Form.Submit message={m.createOrganization} component={SubmitButton} />
        </SubmitBar>
      </Form>
    )
  }
}
export default OrganizationForm
