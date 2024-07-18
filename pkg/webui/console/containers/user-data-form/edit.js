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
import { useDispatch, useSelector } from 'react-redux'
import { useNavigate, useParams } from 'react-router-dom'
import { defineMessages, useIntl } from 'react-intl'

import toast from '@ttn-lw/components/toast'
import PageTitle from '@ttn-lw/components/page-title'
import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import Select from '@ttn-lw/components/select'
import Checkbox from '@ttn-lw/components/checkbox'
import SubmitBar from '@ttn-lw/components/submit-bar'
import SubmitButton from '@ttn-lw/components/submit-button'
import DeleteModalButton from '@ttn-lw/components/delete-modal-button'

import Yup from '@ttn-lw/lib/yup'
import { userId as userIdRegexp } from '@ttn-lw/lib/regexp'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import diff from '@ttn-lw/lib/diff'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import capitalizeMessage from '@ttn-lw/lib/capitalize-message'

import { updateUser, deleteUser } from '@console/store/actions/users'

import { selectSelectedUser } from '@console/store/selectors/users'

const approvalStates = [
  'STATE_REQUESTED',
  'STATE_APPROVED',
  'STATE_REJECTED',
  'STATE_FLAGGED',
  'STATE_SUSPENDED',
]

const m = defineMessages({
  deleteWarning:
    "This will <strong>PERMANENTLY DELETE THIS ACCOUNT</strong> and <strong>LOCK THE USER ID AND EMAIL FOR RE-REGISTRATION</strong>. Associated entities (e.g. gateways, applications and end devices) owned by this user that do not have any other collaborators will become <strong>UNACCESSIBLE</strong> and it will <strong>NOT BE POSSIBLE TO REGISTER ENTITIES WITH THE SAME ID OR EUI's AGAIN</strong>. Make sure you assign new collaborators to such entities if you plan to continue using them.",
  purgeWarning:
    "This will <strong>PERMANENTLY DELETE THIS ACCOUNT</strong>. Associated entities (e.g. gateways, applications and end devices) owned by this user that do not have any other collaborators will become <strong>UNACCESSIBLE</strong> and it will <strong>NOT BE POSSIBLE TO REGISTER ENTITIES WITH THE SAME ID OR EUI's AGAIN</strong>. Make sure you assign new collaborators to such entities if you plan to continue using them.",
  deleteConfirmMessage: "Please type in this user's user ID to confirm.",
  updateSuccess: 'User updated',
  deleteSuccess: 'User deleted',
})

const validationSchema = Yup.object().shape({
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

const UserDataFormEdit = () => {
  const dispatch = useDispatch()
  const navigate = useNavigate()
  const { userId } = useParams()
  const intl = useIntl()
  const user = useSelector(selectSelectedUser)
  const [error, setError] = useState(undefined)

  const wrappedUpdateUser = attachPromise(updateUser)
  const wrappedDeleteUser = attachPromise(deleteUser)

  const initialValues = {
    admin: false,
    ...user,
  }

  const { formatMessage } = intl

  const approvalStateOptions = approvalStates.map(state => ({
    value: state,
    label: capitalizeMessage(formatMessage({ id: `enum:${state}` })),
  }))

  const handleSubmit = useCallback(
    async (vals, { resetForm, setSubmitting }) => {
      const { _validate_email, ...values } = validationSchema.cast(vals)

      if (_validate_email) {
        values.primary_email_address_validated_at = new Date().toISOString()
      }

      setError(undefined)
      try {
        const patch = diff(user, values)
        const submitPatch = Object.keys(patch).length !== 0 ? patch : user
        await dispatch(wrappedUpdateUser(userId, submitPatch))
        resetForm({ values: vals })
        toast({
          title: userId,
          message: m.updateSuccess,
          type: toast.types.SUCCESS,
        })
      } catch (error) {
        setSubmitting(false)
        setError(error)
      }
    },
    [dispatch, user, userId, wrappedUpdateUser],
  )

  const handleDelete = useCallback(
    async shouldPurge => {
      try {
        await dispatch(wrappedDeleteUser(userId, { purge: shouldPurge }))
        toast({
          title: userId,
          message: m.deleteSuccess,
          type: toast.types.SUCCESS,
        })

        navigate('/admin-panel/user-management')
      } catch (error) {
        setError(error)
      }
    },
    [dispatch, navigate, userId, wrappedDeleteUser],
  )

  return (
    <div className="container container--xxl grid">
      <PageTitle title={sharedMessages.userEdit} />
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
            disabled
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
          <SubmitBar>
            <Form.Submit message={sharedMessages.saveChanges} component={SubmitButton} />
            <DeleteModalButton
              message={sharedMessages.userDelete}
              entityId={initialValues?.ids?.user_id}
              entityName={initialValues?.name}
              title={sharedMessages.accountDeleteConfirmation}
              confirmMessage={m.deleteConfirmMessage}
              defaultMessage={m.deleteWarning}
              purgeMessage={m.purgeWarning}
              onApprove={handleDelete}
              shouldConfirm
              mayPurge
            />
          </SubmitBar>
        </Form>
      </div>
    </div>
  )
}

export default UserDataFormEdit
