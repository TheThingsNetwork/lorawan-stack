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

import { getEntity } from '../utils'

import ContentTemplate from './template'

const m = defineMessages({
  title: 'Collaborator of {entityType} added or updated',
  body: 'A collaborator of your {entityType} <code>{entityId}</code> on your network has been added or updated.',
  collaborator: '<b>Collaborator:</b> {collaboratorType} <code>{collaboratorId}</code>',
  link: 'You can view and edit this collaborator <Link>here</Link>.',
  preview:
    'A collaborator of your {entityType} {entityId} on your network has been added or updated. Collaborator: {collaboratorType} {collaboratorId}',
})

const getType = entity => {
  if ('organization_ids' in entity) {
    return 'organization'
  }

  return 'user'
}

const getId = entity => {
  if ('organization_ids' in entity) {
    return entity.organization_ids.organization_id
  } else if ('user_ids' in entity) {
    return entity.user_ids.user_id
  }

  return entity.user_id
}

const CollaboratorChangedIcon = () => <Icon icon="collaborators" className="c-tts-primary" />

const CollaboratorChangedPreview = ({ notificationData }) => {
  const { data, entity_ids } = notificationData
  const { ids } = data

  return (
    <Message
      content={m.preview}
      values={{
        entityType: getEntity(entity_ids),
        entityId: entity_ids[`${getEntity(entity_ids)}_ids`][`${getEntity(entity_ids)}_id`],
        collaboratorType: getType(ids),
        collaboratorId: getId(ids),
      }}
    />
  )
}

CollaboratorChangedPreview.propTypes = {
  notificationData: PropTypes.notificationData.isRequired,
}

const CollaboratorChangedTitle = ({ notificationData }) => {
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

CollaboratorChangedTitle.propTypes = {
  notificationData: PropTypes.notificationData.isRequired,
}

const CollaboratorChanged = ({ notificationData }) => {
  const { data, entity_ids } = notificationData
  const { ids } = data
  const messages = {
    body: m.body,
    entities: m.collaborator,
    action: m.link,
  }
  const values = {
    body: {
      entityType: getEntity(entity_ids),
      entityId: entity_ids[`${getEntity(entity_ids)}_ids`][`${getEntity(entity_ids)}_id`],
      collaboratorType: getType(ids),
      collaboratorId: getId(ids),
    },
    entities: {
      collaboratorType: getType(ids),
      collaboratorId: getId(ids),
    },
    action: {
      Link: msg => (
        <Link
          to={`/${getEntity(entity_ids)}s/${
            entity_ids.application_ids.application_id
          }/collaborators/user/${getId(ids)}`}
        >
          {msg}
        </Link>
      ),
    },
  }
  return <ContentTemplate messages={messages} values={values} />
}

CollaboratorChanged.propTypes = {
  notificationData: PropTypes.notificationData.isRequired,
}

CollaboratorChanged.Title = CollaboratorChangedTitle
CollaboratorChanged.Preview = CollaboratorChangedPreview
CollaboratorChanged.Icon = CollaboratorChangedIcon

export default CollaboratorChanged
