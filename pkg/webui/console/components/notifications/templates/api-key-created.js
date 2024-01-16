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

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import { getEntity } from '../utils'

import ContentTemplate from './template'

const m = defineMessages({
  title: 'A new API key has just been created for your {entityType}',
  body: 'A new API key has just been created for your {entityType} <code>{id}</code> on your network.',
  preview:
    'A new API key has just been created for your {entityType} {id} on your network. API Key ID: {apiKeyId}',
})

const ApiKeyCreatedIcon = () => <Icon icon="key" className="c-tts-primary" />

const ApiKeyCreatedPreview = ({ notificationData }) => {
  const { entity_ids, data } = notificationData
  const { id } = data

  return (
    <Message
      content={m.preview}
      values={{
        entityType: getEntity(entity_ids),
        id: entity_ids[`${getEntity(entity_ids)}_ids`][`${getEntity(entity_ids)}_id`],
        apiKeyId: id,
      }}
    />
  )
}

ApiKeyCreatedPreview.propTypes = {
  notificationData: PropTypes.notificationData.isRequired,
}

const ApiKeyCreatedTitle = ({ notificationData }) => {
  const { entity_ids } = notificationData
  return (
    <Message
      content={m.title}
      values={{
        entityType: getEntity(entity_ids),
      }}
    />
  )
}

ApiKeyCreatedTitle.propTypes = {
  notificationData: PropTypes.notificationData.isRequired,
}

const ApiKeyCreated = ({ notificationData }) => {
  const { entity_ids, data } = notificationData
  const { id, rights } = data
  const messages = {
    body: m.body,
    entities: sharedMessages.apiKeyId,
    action: sharedMessages.viewLink,
  }

  const values = {
    body: {
      entityType: getEntity(entity_ids),
      id: entity_ids[`${getEntity(entity_ids)}_ids`][`${getEntity(entity_ids)}_id`],
    },
    entities: {
      apiKeyId: id,
    },
    action: {
      Link: msg => (
        <Link
          to={`/applications/${
            entity_ids[`${getEntity(entity_ids)}_ids`][`${getEntity(entity_ids)}_id`]
          }/api-keys`}
        >
          {msg}
        </Link>
      ),
    },
  }
  return (
    <ContentTemplate
      messages={messages}
      values={values}
      withList
      listTitle={sharedMessages.rightsList}
      listElement={rights}
    />
  )
}

ApiKeyCreated.propTypes = {
  notificationData: PropTypes.notificationData.isRequired,
}

ApiKeyCreated.Title = ApiKeyCreatedTitle
ApiKeyCreated.Preview = ApiKeyCreatedPreview
ApiKeyCreated.Icon = ApiKeyCreatedIcon

export default ApiKeyCreated
