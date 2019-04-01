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

import Tag from '../../../components/tag'
import TagGroup from '../../../components/tag/group'
import FetchTable from '../fetch-table'

import sharedMessages from '../../../lib/shared-messages'
import style from './api-keys-table.styl'

const m = defineMessages({
  keyId: 'Key ID',
  grantedRights: 'Granted Rights',
})

const formatRight = function (right) {
  return right.split('_')
    .slice(1)
    .map(r => r.charAt(0) + r.slice(1).toLowerCase() )
    .join(' ')
}

const RIGHT_TAG_MAX_WIDTH = 140

const headers = [
  {
    name: 'id',
    displayName: m.keyId,
    width: 35,
  },
  {
    name: 'rights',
    displayName: m.grantedRights,
    width: 45,
    render (rights) {
      const tags = rights.map(r => (
        <Tag
          className={style.rightTag}
          content={formatRight(r)}
          key={r}
        />
      ))

      return (
        <TagGroup
          tagMaxWidth={RIGHT_TAG_MAX_WIDTH}
          tags={tags}
        />
      )
    },
  },
  {
    name: 'name',
    displayName: sharedMessages.name,
    width: 20,
  },
]

export default class ApiKeysTable extends Component {
  render () {
    return (
      <FetchTable
        entity="keys"
        headers={headers}
        addMessage={sharedMessages.addApiKey}
        handlesPagination
        {...this.props}
      />
    )
  }
}

