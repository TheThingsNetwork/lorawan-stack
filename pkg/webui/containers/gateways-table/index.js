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

import sharedMessages from '../../lib/shared-messages'
import FetchTable from '../fetch-table'

import {
  getGatewaysList,
  searchGatewaysList,
} from '../../actions/gateways'

const m = defineMessages({
  add: 'Add Gateway',
  gtwId: 'Gateway ID',
  freqPlan: 'Frequency Plan',
})

const headers = [
  {
    name: 'name',
    displayName: sharedMessages.name,
  },
  {
    name: 'description',
    displayName: sharedMessages.description,
  },
  {
    name: 'antennasCount',
    displayName: sharedMessages.antennas,
    centered: true,
  },
  {
    name: 'frequency_plan',
    displayName: m.freqPlan,
  },
]

export default class GatewaysTable extends React.Component {
  render () {
    return (
      <FetchTable
        entity="gateways"
        addMessage={m.add}
        headers={headers}
        getItemsAction={getGatewaysList}
        searchItemsAction={searchGatewaysList}
        {...this.props}
      />
    )
  }
}
