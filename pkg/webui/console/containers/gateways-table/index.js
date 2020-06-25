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
import { connect } from 'react-redux'

import Status from '@ttn-lw/components/status'

import Message from '@ttn-lw/lib/components/message'

import FetchTable from '@console/containers/fetch-table'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import { checkFromState, mayCreateGateways } from '@console/lib/feature-checks'

import { getGatewaysList } from '@console/store/actions/gateways'

import { selectUserIsAdmin } from '@console/store/selectors/user'
import {
  selectGateways,
  selectGatewaysTotalCount,
  selectGatewaysFetching,
  selectGatewaysError,
} from '@console/store/selectors/gateways'

const m = defineMessages({
  ownedTabTitle: 'Owned gateways',
})

const headers = [
  {
    name: 'ids.gateway_id',
    displayName: sharedMessages.id,
    width: 24,
    sortable: true,
    sortKey: 'gateway_id',
  },
  {
    name: 'name',
    displayName: sharedMessages.name,
    width: 22,
    sortable: true,
  },
  {
    name: 'ids.eui',
    displayName: sharedMessages.gatewayEUI,
    width: 20,
    sortable: true,
    sortKey: 'gateway_eui',
  },
  {
    name: 'frequency_plan_id',
    displayName: sharedMessages.frequencyPlan,
    width: 16,
  },
  {
    name: 'status',
    width: 18,
    displayName: sharedMessages.status,
    render(status) {
      let indicator = 'unknown'
      let label = sharedMessages.unknown

      if (status === 'connected') {
        indicator = 'good'
        label = sharedMessages.connected
      } else if (status === 'disconnected') {
        indicator = 'bad'
        label = sharedMessages.disconnected
      } else if (status === 'other-cluster') {
        indicator = 'unknown'
        label = sharedMessages.otherCluster
      } else if (status === 'unknown') {
        indicator = 'unknown'
        label = sharedMessages.unknown
      }

      return <Status status={indicator} label={label} />
    },
  },
]

const OWNED_TAB = 'owned'
const ALL_TAB = 'all'
const tabs = [
  {
    title: m.ownedTabTitle,
    name: OWNED_TAB,
  },
  {
    title: sharedMessages.allAdmin,
    name: ALL_TAB,
  },
]

class GatewaysTable extends React.Component {
  constructor(props) {
    super(props)

    this.getGatewaysList = params => {
      const { tab, query } = params

      return getGatewaysList(
        params,
        ['name', 'description', 'frequency_plan_id', 'gateway_server_address'],
        { withStatus: true, isSearch: tab === ALL_TAB || query.length > 0 },
      )
    }
  }

  static propTypes = {
    isAdmin: PropTypes.bool.isRequired,
  }

  baseDataSelector(state) {
    return {
      gateways: selectGateways(state),
      totalCount: selectGatewaysTotalCount(state),
      fetching: selectGatewaysFetching(state),
      error: selectGatewaysError(state),
      mayAdd: checkFromState(mayCreateGateways, state),
    }
  }

  render() {
    const { isAdmin, ...rest } = this.props
    return (
      <FetchTable
        entity="gateways"
        addMessage={sharedMessages.addGateway}
        headers={headers}
        getItemsAction={this.getGatewaysList}
        baseDataSelector={this.baseDataSelector}
        tableTitle={<Message content={sharedMessages.gateways} />}
        searchable
        tabs={isAdmin ? tabs : []}
        {...rest}
      />
    )
  }
}

export default connect(state => ({
  isAdmin: selectUserIsAdmin(state),
}))(GatewaysTable)
