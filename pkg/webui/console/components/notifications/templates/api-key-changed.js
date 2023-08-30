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
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { getEntity } from '../utils'

import ContentTemplate from './template'

const m = defineMessages({
  title: 'An API key of your {entityType} on your network has been changed',
  body: 'An API key of your {entityType} <code>{id}</code> on your network has been changed.',
  preview:
    'An API key of your {entityType} "{id}" on your network has been changed. API Key ID: {apiKeyId}',
})

const ApiKeyChangedPreview = ({ notificationData }) => {
  const { entity_ids, data } = notificationData
  const { id } = data

  return (
    <Message
      content={m.preview}
      values={{
        entityType: getEntity(entity_ids),
        id: entity_ids[`${getEntity(entity_ids)}_ids`][`${getEntity(entity_ids)}_id`],
        apiKeyId: id,
        linebreak: <br />,
      }}
    />
  )
}

ApiKeyChangedPreview.propTypes = {
  notificationData: PropTypes.shape({
    data: PropTypes.shape({
      id: PropTypes.string.isRequired,
    }).isRequired,
    entity_ids: PropTypes.shape({}).isRequired,
  }).isRequired,
}

const ApiKeyChangedTitle = ({ notificationData }) => {
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

ApiKeyChangedTitle.propTypes = {
  notificationData: PropTypes.shape({
    entity_ids: PropTypes.shape({}).isRequired,
  }).isRequired,
}

const ApiKeyChanged = ({ notificationData }) => {
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
      code: msg => <code>{msg}</code>,
      b: msg => <b>{msg}</b>,
    },
    entities: {
      b: msg => <b>{msg}</b>,
      code: msg => <code>{msg}</code>,
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

ApiKeyChanged.propTypes = {
  notificationData: PropTypes.shape({
    data: PropTypes.shape({
      id: PropTypes.string.isRequired,
      rights: PropTypes.arrayOf(PropTypes.string).isRequired,
    }).isRequired,
    entity_ids: PropTypes.shape({}).isRequired,
  }).isRequired,
}

ApiKeyChanged.Title = ApiKeyChangedTitle
ApiKeyChanged.Preview = ApiKeyChangedPreview

export default ApiKeyChanged
