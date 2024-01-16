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

import React from 'react'
import { defineMessages } from 'react-intl'

import Link from '@ttn-lw/components/link'
import Icon from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import ContentTemplate from './template'

const m = defineMessages({
  title: 'Your review is required for a newly registered user',
  body: 'A new user just registered on your network.{lineBreak}Since user registration requires admin approval, you need to approve this user before they can login.',
  closing: 'You can review this user <Link>here</Link>.',
  user: '<b>User ID:</b> <code>{userId}</code>{lineBreak}<b>Name:</b> {userName}{lineBreak}<b>Description:</b> {userDescription}{lineBreak}<b>Email Address:</b> {userPrimaryEmailAddress}',
  preview:
    'A new user just registered on your network. Since user registration requires admin approval, you need to approve this user before they can login. User ID: {userId}',
})

const UserRequestedIcon = () => <Icon icon="person_add" className="c-tts-primary" />

const UserRequestedPreview = ({ notificationData }) => {
  const { user } = notificationData.data

  return (
    <Message
      content={m.preview}
      values={{
        userId: user.ids.user_id,
      }}
    />
  )
}

UserRequestedPreview.propTypes = {
  notificationData: PropTypes.notificationData.isRequired,
}

const UserRequestedTitle = () => <Message content={m.title} />

const UserRequested = ({ notificationData }) => {
  const { user } = notificationData.data
  const messages = {
    body: m.body,
    entities: m.user,
    action: m.closing,
  }
  const values = {
    entities: {
      userId: user.ids.user_id,
      userName: user.name,
      userDescription: user.description ?? '—',
      userPrimaryEmailAddress: user.primary_email_address,
    },
    action: {
      Link: msg => <Link to={`/admin-panel/user-management/${user.ids.user_id}`}>{msg}</Link>,
    },
  }
  return <ContentTemplate messages={messages} values={values} />
}

UserRequested.propTypes = {
  notificationData: PropTypes.notificationData.isRequired,
}

UserRequested.Title = UserRequestedTitle
UserRequested.Preview = UserRequestedPreview
UserRequested.Icon = UserRequestedIcon

export default UserRequested
