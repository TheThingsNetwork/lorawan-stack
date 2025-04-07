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

import React, { useCallback } from 'react'
import { defineMessages } from 'react-intl'
import { useParams } from 'react-router-dom'
import { createSelector } from 'reselect'

import FetchTable from '@ttn-lw/containers/fetch-table'

import Message from '@ttn-lw/lib/components/message'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { natsUrl as natsUrlRegexp } from '@console/lib/regexp'

import { getPubsubsList } from '@console/store/actions/pubsubs'

import { selectPubsubs, selectPubsubsTotalCount } from '@console/store/selectors/pubsubs'

const m = defineMessages({
  host: 'Server host',
})

const headers = [
  {
    name: 'ids.pub_sub_id',
    displayName: sharedMessages.id,
    width: 40,
    sortable: true,
  },
  {
    getValue: row => {
      if (row.nats) {
        const res = row.nats.server_url.match(natsUrlRegexp)
        return res ? res[8] : ''
      } else if (row.mqtt) {
        return row.mqtt.server_url
      }
      return ''
    },
    displayName: m.host,
    width: 33,
  },
  {
    name: 'base_topic',
    displayName: sharedMessages.pubsubBaseTopic,
    width: 9,
    sortable: true,
  },
  {
    getValue: row => {
      if (row.nats) {
        return 'NATS'
      } else if (row.mqtt) {
        return 'MQTT'
      }
      return 'Not set'
    },
    displayName: sharedMessages.provider,
    width: 9,
  },
  {
    name: 'format',
    displayName: sharedMessages.format,
    width: 9,
    sortable: true,
  },
]

const getItemPathPrefix = item => `${item.ids.pub_sub_id}`

const PubsubsTable = () => {
  const { appId } = useParams()

  const getItemsAction = useCallback(filters => getPubsubsList(appId, { ...filters }), [appId])

  const baseDataSelector = createSelector(
    [selectPubsubs, selectPubsubsTotalCount],
    (pubsubs, totalCount) => ({
      pubsubs,
      totalCount,
    }),
  )

  return (
    <FetchTable
      entity="pubsubs"
      defaultOrder="-ids.pub_sub_id"
      addMessage={sharedMessages.addPubsub}
      headers={headers}
      getItemsAction={getItemsAction}
      baseDataSelector={baseDataSelector}
      tableTitle={<Message content={sharedMessages.pubsubs} />}
      getItemPathPrefix={getItemPathPrefix}
      handlesSorting
    />
  )
}

export default PubsubsTable
