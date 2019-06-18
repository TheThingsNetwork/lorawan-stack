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

import FetchTable from '../fetch-table'
import Message from '../../../lib/components/message'
import sharedMessages from '../../../lib/shared-messages'
import Icon from '../../../components/icon'

import style from './collaborators-table.styl'

const m = defineMessages({
  id: 'User / Organization ID',
})

const headers = [
  {
    name: 'id',
    displayName: m.id,
  },
  {
    name: 'isUser',
    displayName: sharedMessages.type,
    render (isUser) {
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
    render (rights) {
      for (let i = 0; i < rights.length; i++) {
        if (rights[i].includes('_ALL')) {
          return (
            <Message content={sharedMessages.all} />
          )
        }
      }

      return <span>{rights.length}</span>
    },
  },
]

export default class CollaboratorsTable extends Component {
  render () {
    return (
      <FetchTable
        entity="collaborators"
        headers={headers}
        addMessage={sharedMessages.collaboratorAdd}
        {...this.props}
      />
    )
  }
}
