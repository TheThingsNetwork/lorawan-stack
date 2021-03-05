// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

import React, { Component } from 'react'
import { defineMessages } from 'react-intl'
import { connect } from 'react-redux'

import Icon from '@ttn-lw/components/icon'

import FetchTable from '@ttn-lw/containers/fetch-table'

import Message from '@ttn-lw/lib/components/message'

import { getCollaboratorId } from '@ttn-lw/lib/selectors/id'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import { selectUserId } from '@console/store/selectors/user'

import style from './collaborators-table.styl'

const m = defineMessages({
  id: 'User / Organization ID',
})

@connect(state => ({
  currentUserId: selectUserId(state),
}))
export default class CollaboratorsTable extends Component {
  static propTypes = {
    currentUserId: PropTypes.string.isRequired,
  }

  constructor(props) {
    super(props)
    this.headers = [
      {
        name: 'ids',
        displayName: m.id,
        render: ids => {
          const isUser = 'user_ids' in ids
          const collaboratorId = getCollaboratorId({ ids })
          if (isUser && collaboratorId === props.currentUserId) {
            return (
              <span>
                {collaboratorId}{' '}
                <Message className={style.hint} content={sharedMessages.currentUserIndicator} />
              </span>
            )
          }
          return collaboratorId
        },
      },
      {
        name: 'ids',
        displayName: sharedMessages.type,
        render: ids => {
          const isUser = 'user_ids' in ids
          const icon = isUser ? 'user' : 'organization'

          return (
            <span>
              <Icon icon={icon} className={style.collaboratorIcon} />
              <Message content={isUser ? sharedMessages.user : sharedMessages.organization} />
            </span>
          )
        },
      },
      {
        name: 'rights',
        displayName: sharedMessages.rights,
        render: rights => {
          for (let i = 0; i < rights.length; i++) {
            if (rights[i].includes('_ALL')) {
              return <Message content={sharedMessages.all} />
            }
          }

          return <span>{rights.length}</span>
        },
      },
    ]
  }

  getCollaboratorPathPrefix(collaborator) {
    return `/${'user_ids' in collaborator.ids ? 'user' : 'organization'}/${getCollaboratorId(
      collaborator,
    )}`
  }

  render() {
    return (
      <FetchTable
        entity="collaborators"
        headers={this.headers}
        getItemPathPrefix={this.getCollaboratorPathPrefix}
        addMessage={sharedMessages.addCollaborator}
        tableTitle={<Message content={sharedMessages.collaborators} />}
        {...this.props}
      />
    )
  }
}
