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

import { getEntity } from '../utils'

const m = defineMessages({
  title: 'A collaborator of your {entityType} has been added or updated',
  greeting: 'Dear {recieverName},',
  body: 'A collaborator of your {entityType} <code>{entityId}</code> on your local network has been added or updated.',
  collaborator: '<b>Collaborator:</b> {collaboratorType} <code>{collaboratorId}</code>',
  link: 'You can view and edit this collaborator <Link>here</Link>.',
  preview:
    'A collaborator of your {entityType} {entityId} on your local network has been added or updated. Collaborator: {collaboratorType} {collaboratorId}',
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
        lineBreak: <br />,
      }}
    />
  )
}

CollaboratorChangedPreview.propTypes = {
  notificationData: PropTypes.shape({
    data: PropTypes.shape({
      ids: PropTypes.shape({}).isRequired,
    }).isRequired,
    entity_ids: PropTypes.shape({}).isRequired,
  }).isRequired,
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
  notificationData: PropTypes.shape({
    entity_ids: PropTypes.shape({}).isRequired,
  }).isRequired,
}

const CollaboratorChanged = ({ reciever, notificationData }) => {
  const { data, entity_ids } = notificationData
  const { ids } = data

  return (
    <>
      <Message content={m.greeting} values={{ recieverName: reciever }} component="p" />
      <Message
        content={m.body}
        values={{
          entityType: getEntity(entity_ids),
          entityId: entity_ids[`${getEntity(entity_ids)}_ids`][`${getEntity(entity_ids)}_id`],
          collaboratorType: getType(ids),
          collaboratorId: getId(ids),
          code: msg => <code>{msg}</code>,
        }}
        component="p"
      />
      <Message
        component="p"
        content={m.collaborator}
        values={{
          b: msg => <b>{msg}</b>,
          code: msg => <code>{msg}</code>,
          collaboratorType: getType(ids),
          collaboratorId: getId(ids),
        }}
      />
      <Message
        content={m.link}
        values={{
          Link: msg => (
            <Link
              to={`/applications/${
                entity_ids.application_ids.application_id
              }/collaborators/user/${getId(ids)}`}
            >
              {msg}
            </Link>
          ),
        }}
      />
    </>
  )
}

CollaboratorChanged.propTypes = {
  notificationData: PropTypes.shape({
    data: PropTypes.shape({
      ids: PropTypes.oneOfType([
        PropTypes.shape({
          organization_ids: PropTypes.shape({
            organization_id: PropTypes.string.isRequired,
          }),
        }),
        PropTypes.shape({
          user_ids: PropTypes.shape({
            user_id: PropTypes.string.isRequired,
          }),
        }),
      ]).isRequired,
    }).isRequired,
    entity_ids: PropTypes.shape({
      application_ids: PropTypes.shape({
        application_id: PropTypes.string.isRequired,
      }),
    }).isRequired,
  }).isRequired,
  reciever: PropTypes.string.isRequired,
}

CollaboratorChanged.Title = CollaboratorChangedTitle
CollaboratorChanged.Preview = CollaboratorChangedPreview

export default CollaboratorChanged
