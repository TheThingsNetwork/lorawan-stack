// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

import Icon, { IconOauthClients } from '@ttn-lw/components/icon'
import Link from '@ttn-lw/components/link'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'
import capitalizeMessage from '@ttn-lw/lib/capitalize-message'

import selectAccountUrl from '@console/lib/selectors/app-config'

import ContentTemplate from './template'

const m = defineMessages({
  title: 'Review is required for a new OAuth client',
  body: '{senderType} <code>{id}</code> just registered a new OAuth client under {collaboratorType} <code>{collaboratorId}</code> on your network.{lineBreak}Since {senderTypeMiddle} <code>{id}</code> is not an admin, you need to approve this client before it can be used.',
  clientId: '<b>Client ID:</b> <code>{clientId}</code>',
  link: 'You can approve (or reject) the OAuth client <Link>here</Link>.',
  preview:
    '{senderType} {id} just registered a new OAuth client under {collaboratorType} {collaboratorId} on your network. Since {senderTypeMiddle} {id} is not an admin, you need to approve this client before it can be used. Client ID: {clientId}',
})

const accountUrl = selectAccountUrl()

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

const ClientRequestedIcon = () => <Icon icon={IconOauthClients} className="c-tts-primary" />

const ClientRequestedPreview = ({ notificationData }) => {
  const { data, sender_ids } = notificationData
  const client = 'create_client_request' in data ? data.create_client_request.client : data.client
  const collaborator =
    'create_client_request' in data ? data.create_client_request.collaborator : data.collaborator

  return (
    <Message
      content={m.preview}
      values={{
        senderType: sender_ids ? capitalizeMessage(getType(sender_ids)) : 'An unknown',
        senderTypeMiddle: sender_ids ? getType(sender_ids) : 'the unknown',
        id: sender_ids ? getId(sender_ids) : ' user',
        collaboratorType: getType(collaborator),
        collaboratorId: getId(collaborator),
        clientId: client.ids.client_id,
      }}
    />
  )
}

ClientRequestedPreview.propTypes = {
  notificationData: PropTypes.notificationData.isRequired,
}

const ClientRequestedTitle = () => <Message content={m.title} />

const ClientRequested = ({ notificationData }) => {
  const { data, sender_ids } = notificationData
  const client = 'create_client_request' in data ? data.create_client_request.client : data.client
  const collaborator =
    'create_client_request' in data ? data.create_client_request.collaborator : data.collaborator

  const messages = {
    body: m.body,
    entities: m.clientId,
    action: m.link,
  }
  const values = {
    body: {
      senderType: sender_ids ? capitalizeMessage(getType(sender_ids)) : 'An unknown',
      senderTypeMiddle: sender_ids ? getType(sender_ids) : 'the unknown',
      id: sender_ids ? getId(sender_ids) : ' user',
      collaboratorType: getType(collaborator),
      collaboratorId: getId(collaborator),
    },
    entities: {
      clientId: client.ids.client_id,
    },
    action: {
      Link: msg => (
        <Link.Anchor href={`${accountUrl}/oauth-clients/${client.ids.client_id}`} external>
          {msg}
        </Link.Anchor>
      ),
    },
  }

  return <ContentTemplate messages={messages} values={values} />
}

ClientRequested.propTypes = {
  notificationData: PropTypes.notificationData.isRequired,
}

ClientRequested.Title = ClientRequestedTitle
ClientRequested.Preview = ClientRequestedPreview
ClientRequested.Icon = ClientRequestedIcon

export default ClientRequested
