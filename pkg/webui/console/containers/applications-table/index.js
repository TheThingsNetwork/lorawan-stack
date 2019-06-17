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
import bind from 'autobind-decorator'

import sharedMessages from '../../../lib/shared-messages'
import FetchTable from '../fetch-table'

import { getApplicationsList } from '../../../console/store/actions/applications'

const m = defineMessages({
  all: 'All',
  appId: 'Application ID',
  desc: 'Description',
  empty: 'No items matched your criteria',
})

const tabs = [
  {
    title: m.all,
    name: 'all',
    disabled: true,
  },
]

const headers = [
  {
    name: 'ids.application_id',
    displayName: m.appId,
  },
  {
    name: 'description',
    displayName: m.desc,
  },
]

@bind
export default class ApplicationsTable extends Component {

  baseDataSelector ({ applications }) {
    return applications
  }

  render () {
    return (
      <FetchTable
        entity="applications"
        headers={headers}
        addMessage={sharedMessages.addApplication}
        tableTitle={this.tableTitle}
        getItemsAction={getApplicationsList}
        searchItemsAction={getApplicationsList}
        tabs={tabs}
        baseDataSelector={this.baseDataSelector}
        {...this.props}
      />
    )
  }
}

