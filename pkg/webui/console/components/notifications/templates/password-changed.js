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

import Icon from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import { getEntity } from '../utils'

import ContentTemplate from './template'

const m = defineMessages({
  title: 'Password of user "{entityIds}" changed',
  body: 'The password of your user <code>{entityId}</code> on your network has just been changed.',
  closing: 'If this was not done by you, please contact your administrators as soon as possible.',
})

const PasswordChangedIcon = () => <Icon icon="password" className="c-tts-primary" />

const PasswordChangedPreview = ({ notificationData }) => {
  const { entity_ids } = notificationData

  return (
    <Message
      content={m.body}
      values={{
        entityId: entity_ids.user_ids.user_id,
        code: msg => msg,
      }}
    />
  )
}

PasswordChangedPreview.propTypes = {
  notificationData: PropTypes.notificationData.isRequired,
}

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
  notificationData: PropTypes.notificationData.isRequired,
}

const PasswordChanged = ({ notificationData }) => {
  const { entity_ids } = notificationData
  const messages = {
    body: m.body,
    action: m.closing,
  }
  const values = {
    body: {
      entityId: entity_ids.user_ids.user_id,
    },
  }

  return <ContentTemplate messages={messages} values={values} />
}

PasswordChanged.propTypes = {
  notificationData: PropTypes.notificationData.isRequired,
}

PasswordChanged.Title = PasswordChangedTitle
PasswordChanged.Preview = PasswordChangedPreview
PasswordChanged.Icon = PasswordChangedIcon

export default PasswordChanged
