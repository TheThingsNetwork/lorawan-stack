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

import React, { useCallback, useMemo } from 'react'
import { defineMessages } from 'react-intl'
import { createSelector } from 'reselect'

import FetchTable from '@ttn-lw/containers/fetch-table'

import RoutingPolicy from '@console/components/routing-policy'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import { getPacketBrokerNetworkId } from '@ttn-lw/lib/selectors/id'

import { isValidPolicy } from '@console/lib/packet-broker/utils'

import { getPacketBrokerNetworksList } from '@console/store/actions/packet-broker'

import {
  selectPacketBrokerNetworks,
  selectPacketBrokerNetworksTotalCount,
  selectPacketBrokerNetworksFetching,
  selectPacketBrokerForwarderPolicyById,
  selectPacketBrokerHomeNetworkPolicyById,
  selectPacketBrokerOwnCombinedId,
  selectHomeNetworkDefaultRoutingPolicy,
} from '@console/store/selectors/packet-broker'

const m = defineMessages({
  nonDefaultPolicies: 'Networks with non-default policies',
  search: 'Search by tenant ID or name',
  forwarderPolicy: 'Their routing policy towards us',
  homeNetworkPolicy: 'Our routing policy for them',
})

const headers = [
  {
    name: 'id.net_id',
    displayName: sharedMessages.netId,
    width: 10,
    render: netId => netId.toString(16).padStart(6, '0'),
  },
  {
    name: 'id.tenant_id',
    displayName: sharedMessages.tenantId,
    width: 15,
  },
  {
    name: 'name',
    displayName: sharedMessages.name,
    width: 43,
  },
  {
    name: '_forwarderPolicy',
    displayName: m.forwarderPolicy,
    width: 17,
    render: policy => <RoutingPolicy.Matrix policy={policy} />,
  },
  {
    name: '_homeNetworkPolicy',
    displayName: m.homeNetworkPolicy,
    width: 15,
    align: 'left',
    render: policy => <RoutingPolicy.Matrix policy={policy} />,
  },
]

const NON_DEFAULT_POLICIES = 'non-default'
const ALL_TAB = 'all'
const tabs = [
  {
    title: sharedMessages.all,
    name: ALL_TAB,
  },
  {
    title: m.nonDefaultPolicies,
    name: NON_DEFAULT_POLICIES,
  },
]

const PacketBrokerNetworksTable = () => {
  const getPacketBrokerNetworks = useMemo(
    () => params => {
      const { tab } = params
      const passedParams = { withRoutingPolicy: tab === NON_DEFAULT_POLICIES, ...params }

      return getPacketBrokerNetworksList(passedParams, undefined, { fetchPolicies: true })
    },
    [],
  )

  const baseDataSelector = createSelector(
    [
      selectPacketBrokerOwnCombinedId,
      selectPacketBrokerNetworks,
      selectHomeNetworkDefaultRoutingPolicy,
      selectPacketBrokerForwarderPolicyById,
      selectPacketBrokerHomeNetworkPolicyById,
      selectPacketBrokerNetworksTotalCount,
      selectPacketBrokerNetworksFetching,
    ],
    (
      ownCombinedId,
      networks,
      defaultHomeNetworkRoutingPolicy,
      forwarderPolicy,
      homeNetworkPolicy,
      totalCount,
      fetching,
    ) => {
      const decoratedNetworks = []
      for (const network of networks) {
        const combinedId = getPacketBrokerNetworkId(network)
        if (combinedId === ownCombinedId) {
          continue
        }

        const decoratedNetwork = { ...network }
        decoratedNetwork._forwarderPolicy = forwarderPolicy || {
          uplink: {},
          downlink: {},
        }
        decoratedNetwork._homeNetworkPolicy = isValidPolicy(homeNetworkPolicy)
          ? homeNetworkPolicy
          : defaultHomeNetworkRoutingPolicy
        decoratedNetworks.push(decoratedNetwork)
      }

      return {
        networks: decoratedNetworks,
        totalCount,
        fetching,
        mayAdd: false,
      }
    },
  )

  const getItemPathPrefix = useCallback(network => {
    const netId = network.id.net_id
    const tenantId = network.id.tenant_id

    if (tenantId) {
      return `${netId}/${tenantId}`
    }

    return netId.toString()
  }, [])

  const rowKeySelector = useCallback(
    ({ id }) => `${id.net_id}${'tenant_id' in id ? `/${id.tenant_id}` : ''}`,
    [],
  )

  return (
    <FetchTable
      entity="networks"
      headers={headers}
      getItemsAction={getPacketBrokerNetworks}
      getItemPathPrefix={getItemPathPrefix}
      rowKeySelector={rowKeySelector}
      baseDataSelector={baseDataSelector}
      tabs={tabs}
      searchPlaceholderMessage={m.search}
      searchable
    />
  )
}

export default PacketBrokerNetworksTable
