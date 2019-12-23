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

import React from 'react'
import { defineMessages } from 'react-intl'
import bind from 'autobind-decorator'

import FetchTable from '../fetch-table'
import Message from '../../../lib/components/message'

import sharedMessages from '../../../lib/shared-messages'

import { getPubsubsList } from '../../../console/store/actions/pubsubs'
import {
  selectPubsubs,
  selectPubsubsTotalCount,
  selectPubsubsFetching,
} from '../../../console/store/selectors/pubsubs'
import { natsUrl as natsUrlRegexp } from '../../lib/regexp'
import PropTypes from '../../../lib/prop-types'

const m = defineMessages({
  format: 'Format',
  host: 'Server Host',
})

const headers = [
  {
    name: 'ids.pub_sub_id',
    displayName: sharedMessages.id,
    width: 25,
  },
  {
    getValue(row) {
      if (row.nats) {
        return 'NATS'
      } else if (row.mqtt) {
        return 'MQTT'
      }
      return 'Not set'
    },
    displayName: sharedMessages.provider,
    width: 10,
  },
  {
    name: 'format',
    displayName: m.format,
    width: 10,
  },
  {
    name: 'base_topic',
    displayName: sharedMessages.pubsubBaseTopic,
    width: 15,
  },
  {
    getValue(row) {
      if (row.nats) {
        const res = row.nats.server_url.match(natsUrlRegexp)
        return res ? res[8] : ''
      } else if (row.mqtt) {
        return row.mqtt.server_url
      }
      return ''
    },
    displayName: m.host,
    width: 40,
  },
]

const getItemPathPrefix = item => `/${item.ids.pub_sub_id}`

@bind
export default class PubsubsTable extends React.Component {
  static propTypes = {
    appId: PropTypes.string.isRequired,
  }

  constructor(props) {
    super(props)

    const { appId } = props
    this.getPubsubsList = () => getPubsubsList(appId)
  }

  baseDataSelector(state) {
    return {
      pubsubs: selectPubsubs(state),
      totalCount: selectPubsubsTotalCount(state),
      fetching: selectPubsubsFetching(state),
    }
  }

  render() {
    return (
      <FetchTable
        entity="pubsubs"
        addMessage={sharedMessages.addPubsub}
        headers={headers}
        getItemsAction={this.getPubsubsList}
        baseDataSelector={this.baseDataSelector}
        tableTitle={<Message content={sharedMessages.pubsubs} />}
        getItemPathPrefix={getItemPathPrefix}
        {...this.props}
      />
    )
  }
}
