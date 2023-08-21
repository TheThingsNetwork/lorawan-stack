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

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'
import capitalizeMessage from '@ttn-lw/lib/capitalize-message'

import { getEntity } from '../utils'

const m = defineMessages({
  title: 'The state of your {entityType} has been changed.',
  greeting: 'Dear {recieverName},',
  body: 'The state of the {entityType} <code>{entityId}</code> on your local network has been changed to "{state}".',
  link: 'You can view this <Link>here</Link>.',
})

const EntityStateChangedtPreview = ({ notificationData }) => {
  const { data, entity_ids } = notificationData
  const { formatMessage } = useIntl()

  return (
    <Message
      content={m.body}
      values={{
        entityType: getEntity(entity_ids),
        entityId: entity_ids[`${getEntity(entity_ids)}_ids`][`${getEntity(entity_ids)}_id`],
        state: capitalizeMessage(formatMessage({ id: `enum:${data.state}` })),
        code: msg => msg,
      }}
    />
  )
}

EntityStateChangedtPreview.propTypes = {
  notificationData: PropTypes.shape({
    data: PropTypes.shape({
      state: PropTypes.string.isRequired,
    }).isRequired,
    entity_ids: PropTypes.shape({}).isRequired,
  }).isRequired,
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
  notificationData: PropTypes.shape({
    entity_ids: PropTypes.shape({}).isRequired,
  }).isRequired,
}

const EntityStateChanged = ({ reciever, notificationData }) => {
  const { data, entity_ids } = notificationData
  const { formatMessage } = useIntl()

  return (
    <>
      <Message content={m.greeting} values={{ recieverName: reciever }} component="p" />
      <Message
        content={m.body}
        values={{
          entityType: getEntity(entity_ids),
          entityId: entity_ids[`${getEntity(entity_ids)}_ids`][`${getEntity(entity_ids)}_id`],
          state: capitalizeMessage(formatMessage({ id: `enum:${data.state}` })),
          code: msg => <code>{msg}</code>,
        }}
        component="p"
      />
    </>
  )
}

EntityStateChanged.propTypes = {
  notificationData: PropTypes.shape({
    data: PropTypes.shape({
      state: PropTypes.string.isRequired,
    }).isRequired,
    entity_ids: PropTypes.shape({
      user_ids: PropTypes.shape({
        user_id: PropTypes.string.isRequired,
      }),
    }).isRequired,
  }).isRequired,
  reciever: PropTypes.string.isRequired,
}

EntityStateChanged.Title = EntityStateChangedTitle
EntityStateChanged.Preview = EntityStateChangedtPreview

export default EntityStateChanged
