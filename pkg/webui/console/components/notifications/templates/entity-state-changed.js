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
import { defineMessages, useIntl } from 'react-intl'

import Link from '@ttn-lw/components/link'
import Icon from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'
import capitalizeMessage from '@ttn-lw/lib/capitalize-message'

import { getEntity } from '../utils'

import ContentTemplate from './template'

const m = defineMessages({
  title: 'The state of your {entityType} has been changed.',
  body: 'The state of the {entityType} <code>{entityId}</code> on your network has been changed to "{state}".',
  link: 'You can view this <Link>here</Link>.',
})

const EntityStateChangedIcon = () => <Icon icon="person_add" className="c-tts-primary" />

const EntityStateChangedPreview = ({ notificationData }) => {
  const { data, entity_ids } = notificationData
  const { formatMessage } = useIntl()

  return (
    <Message
      content={m.body}
      values={{
        entityType: getEntity(entity_ids),
        entityId: entity_ids[`${getEntity(entity_ids)}_ids`][`${getEntity(entity_ids)}_id`],
        state: formatMessage({ id: `enum:${data.state}` }),
        code: msg => msg,
      }}
    />
  )
}

EntityStateChangedPreview.propTypes = {
  notificationData: PropTypes.notificationData.isRequired,
}

const EntityStateChangedTitle = ({ notificationData }) => {
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

EntityStateChangedTitle.propTypes = {
  notificationData: PropTypes.notificationData.isRequired,
}

const EntityStateChanged = ({ notificationData }) => {
  const { data, entity_ids } = notificationData
  const { formatMessage } = useIntl()
  const messages = {
    body: m.body,
    action: m.link,
  }
  const values = {
    body: {
      entityType: getEntity(entity_ids),
      entityId: entity_ids[`${getEntity(entity_ids)}_ids`][`${getEntity(entity_ids)}_id`],
      state: capitalizeMessage(formatMessage({ id: `enum:${data.state}` })),
    },
    action: {
      Link: msg => (
        <Link
          to={`/${getEntity(entity_ids)}/${
            entity_ids[`${getEntity(entity_ids)}_ids`][`${getEntity(entity_ids)}_id`]
          }`}
        >
          {msg}
        </Link>
      ),
    },
  }

  return <ContentTemplate messages={messages} values={values} />
}

EntityStateChanged.propTypes = {
  notificationData: PropTypes.notificationData.isRequired,
}

EntityStateChanged.Title = EntityStateChangedTitle
EntityStateChanged.Preview = EntityStateChangedPreview
EntityStateChanged.Icon = EntityStateChangedIcon

export default EntityStateChanged
