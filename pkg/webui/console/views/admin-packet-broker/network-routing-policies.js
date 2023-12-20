// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'

import PacketBrokerNetworksTable from '@console/containers/packet-broker-networks-table'

import sharedMessages from '@ttn-lw/lib/shared-messages'

const NetworkRoutingPoliciesView = () => {
  useBreadcrumbs(
    'admin-panel.packet-broker.routing-configuration.networks',
    <Breadcrumb
      path={'/admin-panel/packet-broker/routing-configuration/networks'}
      content={sharedMessages.networks}
    />,
  )

  return <PacketBrokerNetworksTable />
}

export default NetworkRoutingPoliciesView
