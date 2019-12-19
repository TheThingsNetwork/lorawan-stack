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
import { defineMessages, injectIntl } from 'react-intl'
import * as Yup from 'yup'

import Form from '../../../components/form'
import Input from '../../../components/input'
import Select from '../../../components/select'
import Checkbox from '../../../components/checkbox'
import SubmitBar from '../../../components/submit-bar'
import SubmitButton from '../../../components/submit-button'
import ModalButton from '../../../components/button/modal-button'

import PropTypes from '../../../lib/prop-types'
import sharedMessages from '../../../lib/shared-messages'

const capitalize = str => str.charAt(0).toUpperCase() + str.slice(1)

const approvalStates = [
  'STATE_REQUESTED',
  'STATE_APPROVED',
  'STATE_REJECTED',
  'STATE_FLAGGED',
  'STATE_SUSPENDED',
]

const validationSchema = Yup.object().shape({
  name: Yup.string()
    .min(2, sharedMessages.validateTooShort)
    .max(50, sharedMessages.validateTooLong),
  primary_email_address: Yup.string().email(sharedMessages.validateEmail),
  state: Yup.string().oneOf(approvalStates),
  description: Yup.string().max(2000, sharedMessages.validateTooLong),
})

const m = defineMessages({
  adminLabel: 'Administrator',
  userDescPlaceholder: 'Description for my new user',
  userDescDescription: 'Optional user description; can also be used to save notes about the user',
  userIdPlaceholder: 'jane-doe',
  userNamePlaceholder: 'Jane Doe',
  emailPlaceholder: 'mail@example.com',
  emailAddressDescription:
    'Primary email address used for logging in; this address is not publicly visible',
  modalWarning:
    'Are you sure you want to delete the user "{userId}". This action cannot be undone!',
})

@injectIntl
class UserForm extends React.Component {
  static propTypes = {
    error: PropTypes.error,
    initialValues: PropTypes.shape({
      ids: PropTypes.shape({
        user_id: PropTypes.string.isRequired,
      }).isRequired,
      name: PropTypes.string,
      description: PropTypes.string,
    }).isRequired,
    intl: PropTypes.shape({
      formatMessage: PropTypes.func.isRequired,
    }).isRequired,
    onDelete: PropTypes.func.isRequired,
    onDeleteFailure: PropTypes.func,
    onDeleteSuccess: PropTypes.func,
    onSubmit: PropTypes.func.isRequired,
    onSubmitFailure: PropTypes.func,
    onSubmitSuccess: PropTypes.func,
  }

  static defaultProps = {
    error: '',
    onSubmitFailure: () => null,
    onSubmitSuccess: () => null,
    onDeleteFailure: () => null,
    onDeleteSuccess: () => null,
  }

  @bind
  async handleSubmit(values, { resetForm, setSubmitting }) {
    const { onSubmit, onSubmitSuccess, onSubmitFailure } = this.props
    const castedValues = validationSchema.cast(values)

    try {
      const result = await onSubmit(castedValues)
      onSubmitSuccess(result)
      resetForm(values)
    } catch (error) {
      setSubmitting(false)
      onSubmitFailure(error)
    }
  }

  @bind
  async handleDelete() {
    const { onDelete, onDeleteSuccess, onDeleteFailure } = this.props
    try {
      await onDelete()
      onDeleteSuccess()
    } catch (error) {
      await this.setState({ error })
      onDeleteFailure()
    }
  }

  render() {
    const {
      error,
      initialValues: values,
      intl: { formatMessage },
    } = this.props

    const approvalStateOptions = approvalStates.map(state => ({
      value: state,
      label: capitalize(formatMessage({ id: `enum:${state}` })),
    }))

    const initialValues = {
      admin: false,
      ...values,
    }

    return (
      <Form
        error={error}
        onSubmit={this.handleSubmit}
        initialValues={initialValues}
        validationSchema={validationSchema}
      >
        <Form.Field
          title={sharedMessages.userId}
          name="ids.user_id"
          component={Input}
          disabled
          required
        />
        <Form.Field
          title={sharedMessages.name}
          name="name"
          placeholder={m.userNamePlaceholder}
          component={Input}
        />
        <Form.Field
          title={sharedMessages.description}
          name="description"
          type="textarea"
          placeholder={m.userDescPlaceholder}
          description={m.userDescDescription}
          component={Input}
        />
        <Form.Field
          title={sharedMessages.emailAddress}
          name="primary_email_address"
          placeholder={m.emailPlaceholder}
          description={m.emailAddressDescription}
          component={Input}
          required
        />
        <Form.Field
          title={sharedMessages.state}
          name="state"
          component={Select}
          options={approvalStateOptions}
        />
        <Form.Field
          title={sharedMessages.admin}
          name="admin"
          component={Checkbox}
          label={m.adminLabel}
        />
        <SubmitBar>
          <Form.Submit message={sharedMessages.saveChanges} component={SubmitButton} />
          <ModalButton
            type="button"
            icon="delete"
            danger
            naked
            message={sharedMessages.userDelete}
            modalData={{
              message: {
                values: { userId: initialValues.name || initialValues.ids.user_id },
                ...m.modalWarning,
              },
            }}
            onApprove={this.handleDelete}
          />
        </SubmitBar>
      </Form>
    )
  }
}
export default UserForm
