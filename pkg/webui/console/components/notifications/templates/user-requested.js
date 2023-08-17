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

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

const m = defineMessages({
  title: 'Your review is required for a newly registered user',
  greeting: 'Dear {recieverName},',
  body: 'A new user just registered on your network.{lineBreak}Since user registration requires admin approval, you need to approve this user before they can login.',
  closing: 'You can review this user <Link>here</Link>.',
  userId: '<b>User ID:</b> <code>{userId}</code>',
  userName: '<b>Name:</b> {userName}',
  userDescription: '<b>Description:</b> {userDescription}',
  userEmail: '<b>Email Address:</b> {userPrimaryEmailAddress}',
  preview:
    'A new user just registered on your network. Since user registration requires admin approval, you need to approve this user before they can login. User ID: {userId}',
})

const UserRequestedPreview = ({ notificationData }) => {
  const { user } = notificationData.data

  return (
    <Message
      content={m.preview}
      values={{
        lineBreak: <br />,
        userId: user.ids.user_id,
      }}
    />
  )
}

UserRequestedPreview.propTypes = {
  notificationData: PropTypes.shape({
    data: PropTypes.shape({
      user: PropTypes.shape({
        ids: PropTypes.shape({
          user_id: PropTypes.string,
        }),
      }),
    }).isRequired,
  }).isRequired,
}

const UserRequestedTitle = () => <Message content={m.title} />

const UserRequested = ({ reciever, notificationData }) => {
  const { user } = notificationData.data

  return (
    <>
      <Message content={m.greeting} values={{ recieverName: reciever }} component="p" />
      <Message
        content={m.body}
        values={{
          lineBreak: <br />,
        }}
        component="p"
      />
      {'ids' in user && (
        <Message
          content={m.userId}
          values={{
            b: msg => <b>{msg}</b>,
            code: msg => <code>{msg}</code>,
            userId: user.ids.user_id,
          }}
          component="p"
        />
      )}
      {'name' in user && (
        <Message
          content={m.userName}
          values={{ b: msg => <b>{msg}</b>, code: msg => <code>{msg}</code>, userName: user.name }}
          component="p"
        />
      )}
      {'description' in user && (
        <Message
          content={m.userDescription}
          values={{
            b: msg => <b>{msg}</b>,
            code: msg => <code>{msg}</code>,
            userDescription: user.description,
          }}
          component="p"
        />
      )}
      {'primary_email_address' in user && (
        <Message
          content={m.userEmail}
          values={{
            b: msg => <b>{msg}</b>,
            code: msg => <code>{msg}</code>,
            userPrimaryEmailAddress: user.primary_email_address,
          }}
          component="p"
        />
      )}
      <Message
        content={m.closing}
        values={{
          Link: msg => <Link to={`/admin-panel/user-management/${user.ids.user_id}`}>{msg}</Link>,
        }}
        component="p"
      />
    </>
  )
}

UserRequested.propTypes = {
  notificationData: PropTypes.shape({
    data: PropTypes.shape({
      user: PropTypes.shape({
        ids: PropTypes.shape({
          user_id: PropTypes.string,
        }),
        name: PropTypes.string,
        description: PropTypes.string,
        primary_email_address: PropTypes.string,
      }).isRequired,
    }).isRequired,
  }).isRequired,
  reciever: PropTypes.string.isRequired,
}

UserRequested.Title = UserRequestedTitle
UserRequested.Preview = UserRequestedPreview

export default UserRequested
