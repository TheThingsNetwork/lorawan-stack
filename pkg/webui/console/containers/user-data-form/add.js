// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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
import { useSelector, useDispatch } from 'react-redux'
import { useNavigate } from 'react-router-dom'
import { useIntl } from 'react-intl'

import PageTitle from '@ttn-lw/components/page-title'
import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import Select from '@ttn-lw/components/select'
import Checkbox from '@ttn-lw/components/checkbox'
import SubmitBar from '@ttn-lw/components/submit-bar'
import SubmitButton from '@ttn-lw/components/submit-button'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import Yup from '@ttn-lw/lib/yup'
import createPasswordValidationSchema from '@ttn-lw/lib/create-password-validation-schema'
import { userId as userIdRegexp } from '@ttn-lw/lib/regexp'
import capitalizeMessage from '@ttn-lw/lib/capitalize-message'

import { createUser } from '@console/store/actions/users'

import { selectPasswordRequirements } from '@console/store/selectors/identity-server'

const approvalStates = [
  'STATE_REQUESTED',
  'STATE_APPROVED',
  'STATE_REJECTED',
  'STATE_FLAGGED',
  'STATE_SUSPENDED',
]

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

const UserDataFormAdd = () => {
  const dispatch = useDispatch()
  const navigate = useNavigate()
  const intl = useIntl()
  const [error, setError] = useState(undefined)

  const passwordRequirements = useSelector(selectPasswordRequirements)
  const validationSchema = baseValidationSchema.concat(
    createPasswordValidationSchema(passwordRequirements),
  )
  const createUserAction = useCallback(values => dispatch(createUser(values)), [dispatch])

  const { formatMessage } = intl

  const approvalStateOptions = approvalStates.map(state => ({
    value: state,
    label: capitalizeMessage(formatMessage({ id: `enum:${state}` })),
  }))

  const initialValues = {
    ids: {
      user_id: '',
    },
    name: '',
    description: '',
    primary_email_address: '',
    state: '',
    admin: false,
  }

  const handleSubmit = useCallback(
    async (vals, { resetForm, setSubmitting }) => {
      const { _validate_email, ...values } = validationSchema.cast(vals)

      if (_validate_email) {
        values.primary_email_address_validated_at = new Date().toISOString()
      }

      setError(undefined)
      try {
        await createUserAction(values)
        resetForm({ values: vals })
        navigate('/admin-panel/user-management')
      } catch (error) {
        setSubmitting(false)
        setError(error)
      }
    },
    [createUserAction, navigate, validationSchema],
  )

  return (
    <div className="container container--xl grid">
      <PageTitle title={sharedMessages.userAdd} />
      <div className="item-12">
        <Form
          error={error}
          onSubmit={handleSubmit}
          initialValues={initialValues}
          validationSchema={validationSchema}
        >
          <Form.Field
            title={sharedMessages.userId}
            name="ids.user_id"
            component={Input}
            autoFocus
            required
          />
          <Form.Field
            title={sharedMessages.name}
            name="name"
            placeholder={sharedMessages.userNamePlaceholder}
            component={Input}
          />
          <Form.Field
            title={sharedMessages.description}
            name="description"
            type="textarea"
            placeholder={sharedMessages.userDescription}
            description={sharedMessages.userDescDescription}
            component={Input}
          />
          <Form.Field
            title={sharedMessages.emailAddress}
            name="primary_email_address"
            placeholder={sharedMessages.emailPlaceholder}
            description={sharedMessages.emailAddressDescription}
            component={Input}
            required
          />
          <Form.Field
            title={sharedMessages.state}
            name="state"
            component={Select}
            options={approvalStateOptions}
            required
          />
          <Form.Field
            name="_validate_email"
            component={Checkbox}
            label={sharedMessages.emailAddressValidation}
            description={sharedMessages.emailAddressValidationDescription}
          />
          <Form.Field
            name="admin"
            component={Checkbox}
            label={sharedMessages.grantAdminStatus}
            description={sharedMessages.adminDescription}
          />
          <Form.Field
            title={sharedMessages.password}
            component={Input}
            name="password"
            type="password"
            autoComplete="new-password"
            required
          />
          <Form.Field
            title={sharedMessages.confirmPassword}
            component={Input}
            name="confirmPassword"
            type="password"
            autoComplete="new-password"
            required
          />
          <SubmitBar>
            <Form.Submit message={sharedMessages.userAdd} component={SubmitButton} />
          </SubmitBar>
        </Form>
      </div>
    </div>
  )
}

export default UserDataFormAdd
