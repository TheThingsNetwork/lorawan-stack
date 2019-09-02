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

import sharedMessages from '../../../lib/shared-messages'
import Message from '../../../lib/components/message'
import PropTypes from '../../../lib/prop-types'

import FetchTable from '../fetch-table'

import { getOrganizationsList } from '../../../console/store/actions/organizations'
import {
  selectOrganizations,
  selectOrganizationsTotalCount,
  selectOrganizationsFetching,
  selectOrganizationsError,
} from '../../store/selectors/organizations'

const headers = [
  {
    name: 'ids.organization_id',
    displayName: sharedMessages.id,
    width: 25,
  },
  {
    name: 'name',
    displayName: sharedMessages.name,
    width: 25,
  },
  {
    name: 'description',
    displayName: sharedMessages.description,
    width: 50,
  },
]

export default class OrganizationsTable extends Component {
  static propTypes = {
    pageSize: PropTypes.number.isRequired,
  }

  constructor(props) {
    super(props)

    this.getOrganizationsList = params => getOrganizationsList(params, ['name', 'description'])
  }

  baseDataSelector(state) {
    return {
      organizations: selectOrganizations(state),
      totalCount: selectOrganizationsTotalCount(state),
      fetching: selectOrganizationsFetching(state),
      error: selectOrganizationsError(state),
    }
  }

  render() {
    const { pageSize } = this.props

    return (
      <FetchTable
        entity="organizations"
        headers={headers}
        addMessage={sharedMessages.addOrganization}
        tableTitle={<Message content={sharedMessages.organizations} />}
        getItemsAction={this.getOrganizationsList}
        searchItemsAction={this.getOrganizationsList}
        baseDataSelector={this.baseDataSelector}
        pageSize={pageSize}
      />
    )
  }
}
