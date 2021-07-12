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

import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import Select from '@ttn-lw/components/select'
import Checkbox from '@ttn-lw/components/checkbox'
import SubmitBar from '@ttn-lw/components/submit-bar'
import SubmitButton from '@ttn-lw/components/submit-button'
import ModalButton from '@ttn-lw/components/button/modal-button'

import Yup from '@ttn-lw/lib/yup'
import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import createPasswordValidationSchema from '@ttn-lw/lib/create-password-validation-schema'

import { userId as userIdRegexp } from '@console/lib/regexp'

const capitalize = str => str.charAt(0).toUpperCase() + str.slice(1)

const approvalStates = [
  'STATE_REQUESTED',
  'STATE_APPROVED',
  'STATE_REJECTED',
  'STATE_FLAGGED',
  'STATE_SUSPENDED',
]

const m = defineMessages({
  adminRights: 'Admin rights',
  adminLabel: 'Grant this user admin status',
  userDescPlaceholder: 'Description for my new user',
  userDescDescription: 'Optional user description; can also be used to save notes about the user',
  userIdPlaceholder: 'jane-doe',
  userNamePlaceholder: 'Jane Doe',
  emailPlaceholder: 'mail@example.com',
  emailAddressDescription:
    'Primary email address used for logging in; this address is not publicly visible',
  emailAddressValidation: 'Treat email address as validated',
  emailAddressValidationDescription:
    'Enable this option if you do not need this user to validate the email address',
  modalWarning:
    'Are you sure you want to delete the user "{userId}". This action cannot be undone and it will not be possible to reuse the user ID.',
})

const baseValidationSchema = Yup.object().shape({
  ids: Yup.object().shape({
    user_id: Yup.string()
      .min(2, Yup.passValues(sharedMessages.validateTooShort))
      .max(36, Yup.passValues(sharedMessages.validateTooLong))
      .matches(userIdRegexp, Yup.passValues(sharedMessages.validateIdFormat))
      .required(sharedMessages.validateRequired),
  }),
  name: Yup.string()
    .min(2, Yup.passValues(sharedMessages.validateTooShort))
    .max(50, Yup.passValues(sharedMessages.validateTooLong)),
  primary_email_address: Yup.string()
    .email(sharedMessages.validateEmail)
    .required(sharedMessages.validateRequired),
  state: Yup.string()
    .oneOf(approvalStates, sharedMessages.validateRequired)
    .required(sharedMessages.validateRequired),
  description: Yup.string().max(2000, Yup.passValues(sharedMessages.validateTooLong)),
})

@injectIntl
class UserForm extends React.Component {
  constructor(props) {
    super(props)

    const { update, passwordRequirements } = props
    this.validationSchema = update
      ? baseValidationSchema
      : baseValidationSchema.concat(createPasswordValidationSchema(passwordRequirements))
    this.state = {
      error: '',
    }
  }

  static propTypes = {
    error: PropTypes.error,
    initialValues: PropTypes.shape({
      ids: PropTypes.shape({
        user_id: PropTypes.string.isRequired,
      }).isRequired,
      name: PropTypes.string,
      description: PropTypes.string,
    }),
    intl: PropTypes.shape({
      formatMessage: PropTypes.func.isRequired,
    }).isRequired,
    onDelete: PropTypes.func,
    onDeleteFailure: PropTypes.func,
    onDeleteSuccess: PropTypes.func,
    onSubmit: PropTypes.func.isRequired,
    onSubmitFailure: PropTypes.func,
    onSubmitSuccess: PropTypes.func,
    passwordRequirements: PropTypes.passwordRequirements,
    update: PropTypes.bool,
  }

  static defaultProps = {
    update: false,
    error: '',
    initialValues: {
      ids: { user_id: '' },
      name: '',
      primary_email_address: '',
      state: '',
      description: '',
      password: '',
      confirmPassword: '',
    },
    onSubmitFailure: () => null,
    onSubmitSuccess: () => null,
    onDelete: () => null,
    onDeleteFailure: () => null,
    onDeleteSuccess: () => null,
    passwordRequirements: {},
  }

  @bind
  async handleSubmit(vals, { resetForm, setSubmitting }) {
    const { onSubmit, onSubmitSuccess, onSubmitFailure } = this.props
    const { _validate_email, ...values } = this.validationSchema.cast(vals)

    if (_validate_email) {
      values.primary_email_address_validated_at = new Date().toISOString()
    }

    await this.setState({ error: '' })
    try {
      const result = await onSubmit(values)
      resetForm({ values: vals })
      onSubmitSuccess(result)
    } catch (error) {
      setSubmitting(false)
      this.setState({ error })
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
      update,
      error: passedError,
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

    const { error: submitError } = this.state

    const error = passedError || submitError

    return (
      <Form
        error={error}
        onSubmit={this.handleSubmit}
        initialValues={initialValues}
        validationSchema={this.validationSchema}
      >
        <Form.Field
          title={sharedMessages.userId}
          name="ids.user_id"
          component={Input}
          disabled={update}
          autoFocus={!update}
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
          name="_validate_email"
          component={Checkbox}
          label={m.emailAddressValidation}
          description={m.emailAddressValidationDescription}
        />
        <Form.Field
          title={sharedMessages.state}
          name="state"
          component={Select}
          options={approvalStateOptions}
          required
        />
        <Form.Field title={m.adminRights} name="admin" component={Checkbox} label={m.adminLabel} />
        {!update && (
          <Form.Field
            title={sharedMessages.password}
            component={Input}
            name="password"
            type="password"
            autoComplete="new-password"
            required
          />
        )}
        {!update && (
          <Form.Field
            title={sharedMessages.confirmPassword}
            component={Input}
            name="confirmPassword"
            type="password"
            autoComplete="new-password"
            required
          />
        )}
        <SubmitBar>
          <Form.Submit
            message={update ? sharedMessages.saveChanges : sharedMessages.userAdd}
            component={SubmitButton}
          />
          {update && (
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
          )}
        </SubmitBar>
      </Form>
    )
  }
}
export default UserForm
