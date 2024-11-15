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
import classNames from 'classnames'

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

import { selectUserId, selectUserIsAdmin } from '@console/store/selectors/logout'
import { selectUserById } from '@console/store/selectors/users'

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
  CLIENT_REQUESTED: 'New OAuth client requested',
  USER_REQUESTED: 'New user requested to join',
  API_KEY_CREATED: 'API Key created for an entity',
  API_KEY_CHANGED: 'API key changed in an entity',
  COLLABORATOR_CHANGED: 'Collaborator created or changed in an entity',
  ENTITY_STATE_CHANGED: 'State of an entity has changed',
  INVITATION: 'User invitation has been sent',
  PASSWORD_CHANGED: 'Password was changed',
  errorNotification:
    "Admins can't unsubscribe from all email notifications, since there are notifications that require admin action.",
  requiresAdminAction: "<i>Requires admin action, can't be unselected</i>",
  unsubscribeFromEverything: 'Turn off all email notifications',
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

  const cbs = allNotificationTypes.map((type, index) => (
    <div key={type}>
      <Checkbox
        name={type}
        disabled={isUnsubscribeAll || isAdminNotificationType(type)}
        value={isAdminNotificationType(type)}
        label={m[type]}
        className={classNames('mb-0', { 'mt-cs-s': index !== 0, 'mt-0': index === 0 })}
      />
      {isAdminNotificationType(type) && (
        <Message
          className={style.requiresAdminAction}
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
        <hr className="m-vert-cs-xl" />
        <label className="d-flex j-between al-center gap-cs-m mt-cs-m">
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
          <div className="mt-cs-xs">
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
  const userId = useSelector(selectUserId)
  const user = useSelector(state => selectUserById(state, userId))
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
