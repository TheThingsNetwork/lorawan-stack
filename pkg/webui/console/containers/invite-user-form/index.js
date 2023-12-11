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
import { useNavigate } from 'react-router-dom'
import * as Yup from 'yup'
import { useDispatch } from 'react-redux'

import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import SubmitBar from '@ttn-lw/components/submit-bar'
import SubmitButton from '@ttn-lw/components/submit-button'

import Message from '@ttn-lw/lib/components/message'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import { sendInvite } from '@console/store/actions/users'

const m = defineMessages({
  invitationsDescription:
    'You can invite users to this network by providing an email address. The person will then get an email with instructions on how to join your network.',
})

const validationSchema = Yup.object().shape({
  email: Yup.string().email(sharedMessages.validateEmail).required(sharedMessages.validateRequired),
})

const InviteForm = () => {
  const navigate = useNavigate()
  const dispatch = useDispatch()

  const onSubmitSuccess = useCallback(() => navigate(`/admin-panel/user-management`), [navigate])

  const [error, setError] = useState()
  const handleSubmit = React.useCallback(
    async (values, { resetForm, setSubmitting }) => {
      setError(undefined)
      try {
        const result = await dispatch(attachPromise(sendInvite(values)))
        resetForm({ values })
        onSubmitSuccess(result)
      } catch (error) {
        setSubmitting(false)
        setError({ error })
      }
    },
    [dispatch, onSubmitSuccess],
  )

  const initialValues = {
    email: '',
  }

  return (
    <>
      <Message content={m.invitationsDescription} component="p" />
      <hr className="mb-ls-m" />
      <Form
        error={error}
        onSubmit={handleSubmit}
        initialValues={initialValues}
        validationSchema={validationSchema}
      >
        <Form.Field
          title={sharedMessages.emailAddress}
          component={Input}
          name="email"
          placeholder={sharedMessages.emailPlaceholder}
          required
        />
        <SubmitBar>
          <Form.Submit message={sharedMessages.invite} component={SubmitButton} />
        </SubmitBar>
      </Form>
    </>
  )
}

export default InviteForm
