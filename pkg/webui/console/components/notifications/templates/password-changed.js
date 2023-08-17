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

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import { getEntity } from '../utils'

const m = defineMessages({
  title: 'The password of your user "{entityIds}" has just been changed.',
  greeting: 'Dear {recieverName},',
  body: 'The password of your user <code>{entityId}</code> on your network has just been changed.',
  closing: 'If this was not done by you, please contact your administrators as soon as possible.',
})

const PasswordChangedTitle = ({ notificationData }) => {
  const { entity_ids } = notificationData

  return (
    <Message
      content={m.title}
      values={{
        entityIds: entity_ids[`${getEntity(entity_ids)}_ids`][`${getEntity(entity_ids)}_id`],
      }}
    />
  )
}

PasswordChangedTitle.propTypes = {
  notificationData: PropTypes.shape({
    entity_ids: PropTypes.shape({}).isRequired,
  }).isRequired,
}

const PasswordChanged = ({ reciever, notificationData }) => {
  const { entity_ids } = notificationData

  return (
    <>
      <Message content={m.greeting} values={{ recieverName: reciever }} component="p" />
      <Message
        content={m.body}
        values={{
          entityId: entity_ids.user_ids.user_id,
          code: msg => <code>{msg}</code>,
        }}
        component="p"
      />
      <Message content={m.closing} component="p" />
    </>
  )
}

PasswordChanged.propTypes = {
  notificationData: PropTypes.shape({
    entity_ids: PropTypes.shape({
      user_ids: PropTypes.shape({
        user_id: PropTypes.string.isRequired,
      }),
    }).isRequired,
  }).isRequired,
  reciever: PropTypes.string.isRequired,
}

PasswordChanged.Title = PasswordChangedTitle

export default PasswordChanged
