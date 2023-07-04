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

import Tag from '@ttn-lw/components/tag'
import TagGroup from '@ttn-lw/components/tag/group'

import FetchTable from '@ttn-lw/containers/fetch-table'

import DateTime from '@ttn-lw/lib/components/date-time'
import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import style from './api-keys-table.styl'

const m = defineMessages({
  keyId: 'Key ID',
  grantedRights: 'Granted Rights',
})

const RIGHT_TAG_MAX_WIDTH = 160

const ApiKeysTable = props => {
  const { pageSize, baseDataSelector, getItemsAction } = props
  const intl = useIntl()

  const headers = [
    {
      name: 'id',
      displayName: m.keyId,
      width: 20,
      sortKey: 'api_key_id',
      render: id => <span className={style.keyId}>{id}</span>,
      sortable: true,
    },
    {
      name: 'name',
      displayName: sharedMessages.name,
      width: 20,
      sortable: true,
    },
    {
      name: 'rights',
      displayName: m.grantedRights,
      width: 50,
      render: (rights = []) => {
        if (rights.length === 0) {
          return <Message className={style.none} content={sharedMessages.none} lowercase />
        }
        const tags = rights.map(r => {
          let rightLabel = intl.formatMessage({ id: `enum:${r}` })
          rightLabel = rightLabel.charAt(0).toUpperCase() + rightLabel.slice(1)
          return <Tag className={style.rightTag} content={rightLabel} key={r} />
        })

        return <TagGroup tagMaxWidth={RIGHT_TAG_MAX_WIDTH} tags={tags} />
      },
    },
    {
      name: 'created_at',
      displayName: sharedMessages.createdAt,
      sortable: true,
      width: 10,
      render: date => <DateTime.Relative value={date} />,
    },
  ]

  return (
    <FetchTable
      entity="keys"
      defaultOrder="-created_at"
      headers={headers}
      addMessage={sharedMessages.addApiKey}
      pageSize={pageSize}
      baseDataSelector={baseDataSelector}
      getItemsAction={getItemsAction}
      tableTitle={<Message content={sharedMessages.apiKeys} />}
    />
  )
}

ApiKeysTable.propTypes = {
  baseDataSelector: PropTypes.func.isRequired,
  getItemsAction: PropTypes.func.isRequired,
  pageSize: PropTypes.number,
}

ApiKeysTable.defaultProps = {
  pageSize: undefined,
}

export default ApiKeysTable
