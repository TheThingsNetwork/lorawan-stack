// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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
import { defineMessages } from 'react-intl'

import Form, { useFormContext } from '@ttn-lw/components/form'
import Checkbox from '@ttn-lw/components/checkbox'
import toast from '@ttn-lw/components/toast'
import Switch from '@ttn-lw/components/switch'
import SubmitButton from '@ttn-lw/components/submit-button'
import Button from '@ttn-lw/components/button'
import SubmitBar from '@ttn-lw/components/submit-bar'
import Icon, { IconAlertTriangle } from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'

import Yup from '@ttn-lw/lib/yup'
import diff from '@ttn-lw/lib/diff'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { updateUser } from '@console/store/actions/user'

import { selectUser, selectUserIsAdmin } from '@console/store/selectors/logout'

import style from './email-notifications-form.styl'

const NOTIFICATION_TYPES = [
  'API_KEY_CREATED',
  'API_KEY_CHANGED',
  'COLLABORATOR_CHANGED',
  'ENTITY_STATE_CHANGED',
  'PASSWORD_CHANGED',
]
const ADMIN_NOTIFICATION_TYPES = ['CLIENT_REQUESTED', 'USER_REQUESTED', 'INVITATION']

const validationSchema = Yup.object().shape({
  email_notification_preferences: Yup.object().shape({ types: Yup.array() }).nullable(),
})

const encodePreferences = value => {
  const types = Object.keys(value).filter(key => value[key] === true)
  return value ? types : []
}
const decodePreferences = value => {
  if (value) {
    const types = value.reduce((n, i) => {
      n[i] = true
      return n
    }, {})

    return types
  }
  return {}
}

const m = defineMessages({
  CLIENT_REQUESTED: 'OAuth client requested',
  USER_REQUESTED: 'New user requested',
  API_KEY_CREATED: 'API key created',
  API_KEY_CHANGED: 'API key changed',
  COLLABORATOR_CHANGED: 'Collaborator created or changed',
  ENTITY_STATE_CHANGED: 'Entity state changed',
  INVITATION: 'User invitation',
  PASSWORD_CHANGED: 'Password was changed',
  API_KEY_CREATED_DESCRIPTION: 'Receive an email when an API has been created for an entity.',
  API_KEY_CHANGED_DESCRIPTION: 'Receive an email when an API key has been changed in an entity.',
  CLIENT_REQUESTED_DESCRIPTION: 'Receive an email when a new OAuth client has been requested.',
  COLLABORATOR_CHANGED_DESCRIPTION:
    'Receive an email when a collaborator has been changed in an entity.',
  ENTITY_STATE_CHANGED_DESCRIPTION: 'Receive an email when the state of an entity has changed.',
  INVITATION_DESCRIPTION: 'Receive an email when an invitation has been sent.',
  PASSWORD_CHANGED_DESCRIPTION: 'Receive an email when a password has been changed.',
  USER_REQUESTED_DESCRIPTION: 'Receive an email when a user has requested to join.',
  errorNotification:
    "Admins can't unsubscribe from all email notifications, since there are notifications that require admin action.",
  requiresAdminAction: "<i>Requires admin action, can't be unselected</i>",
  unsubscribeFromEverything: 'Unsubscribe from everything',
  unsubscribeDescription: 'You will continue to receive notifications in the console.',
  discardChanges: 'Discard changes',
  updateEmailPreferences: 'Updated email preferences',
})

const InnerForm = initialValues => {
  const { values, setFieldValue, resetForm } = useFormContext()
  const isAdmin = useSelector(selectUserIsAdmin)
  const [isUnsubscribeAll, setIsUnsubscribeAll] = useState(
    values.email_notification_preferences.types.length === 0,
  )
  const [showErrorNotification, setShowErrorNotification] = useState(false)
  const allNotificationTypes = [...NOTIFICATION_TYPES, ...ADMIN_NOTIFICATION_TYPES]
  const isAdminNotificationType = useCallback(
    type => ADMIN_NOTIFICATION_TYPES.includes(type) && isAdmin,
    [isAdmin],
  )

  const handleUnsubscribeAll = useCallback(
    checked => {
      if (!isAdmin) {
        const types = checked ? [] : values.email_notification_preferences.types
        setFieldValue('email_notification_preferences.types', types)
        setIsUnsubscribeAll(checked)
      } else {
        setShowErrorNotification(true)
      }
    },
    [setFieldValue, values, isAdmin],
  )

  const handleDiscardChanges = useCallback(() => {
    resetForm(initialValues)
  }, [resetForm, initialValues])

  const cbs = allNotificationTypes.map(type => (
    <div key={type}>
      <Checkbox
        name={type}
        disabled={isAdminNotificationType(type)}
        value={isAdminNotificationType(type)}
        label={m[type]}
        className="mb-0 mt-cs-s"
      />
      <Message
        className="c-text-neutral-light w-full ml-cs-l"
        component="div"
        content={m[`${type}_DESCRIPTION`]}
      />
      {isAdminNotificationType(type) && (
        <Message
          className="c-text-neutral-light w-full ml-cs-l"
          component="div"
          content={m.requiresAdminAction}
          values={{ i: str => <i key="bold">{str}</i> }}
        />
      )}
    </div>
  ))

  return (
    <>
      <div className="border-regular p-cs-xl br-xs mt-cs-xxl">
        <Form.Field
          name="email_notification_preferences.types"
          encode={encodePreferences}
          decode={decodePreferences}
          component={Checkbox.Group}
        >
          {cbs}
        </Form.Field>
        <hr />
        <label className="d-flex j-between al-center mt-cs-m">
          <div>
            <Message content={m.unsubscribeFromEverything} />
            <Message
              className="c-text-neutral-light w-full"
              component="div"
              content={m.unsubscribeDescription}
            />
          </div>
          <Switch onChange={handleUnsubscribeAll} checked={isUnsubscribeAll} />
        </label>
        {showErrorNotification && (
          <div>
            <Icon icon={IconAlertTriangle} small className={style.warningIcon} />
            <Message content={m.errorNotification} className="c-text-warning-normal" />
          </div>
        )}
      </div>
      <SubmitBar>
        <Form.Submit component={SubmitButton} message={sharedMessages.saveChanges} />
        <Button type="button" secondary message={m.discardChanges} onClick={handleDiscardChanges} />
      </SubmitBar>
    </>
  )
}

const EmailNotificationsForm = () => {
  const dispatch = useDispatch()
  const [error, setError] = useState()
  const user = useSelector(selectUser)
  const isAdmin = useSelector(selectUserIsAdmin)
  const userEmailNotifications = user.email_notification_preferences

  const initialValues = {
    email_notification_preferences: {
      types: isAdmin
        ? [...ADMIN_NOTIFICATION_TYPES, ...(userEmailNotifications?.types || [])]
        : userEmailNotifications?.types || [],
    },
  }

  const handleSubmit = useCallback(
    async (values, { resetForm, setSubmitting }) => {
      setError(undefined)
      const patch = diff(user.email_notification_preferences, values)
      try {
        await dispatch(attachPromise(updateUser({ id: user.ids.user_id, patch })))
        toast({
          title: sharedMessages.success,
          message: m.updateEmailPreferences,
          type: toast.types.SUCCESS,
        })
      } catch (error) {
        setError(error)
        setSubmitting(false)
        resetForm(initialValues)
      }
    },
    [dispatch, user, initialValues],
  )

  return (
    <Form
      initialValues={initialValues}
      validationSchema={validationSchema}
      onSubmit={handleSubmit}
      error={error}
    >
      <InnerForm initialValues={initialValues} />
    </Form>
  )
}

export default EmailNotificationsForm
