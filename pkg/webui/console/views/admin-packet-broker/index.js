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
import { Switch, Route } from 'react-router'

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import Breadcrumbs from '@ttn-lw/components/breadcrumbs'

import NotFoundRoute from '@ttn-lw/lib/components/not-found-route'
import RequireRequest from '@ttn-lw/lib/components/require-request'

import Require from '@console/lib/components/require'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import { mayConfigurePacketBroker } from '@console/lib/feature-checks'

import { getPacketBrokerInfo } from '@console/store/actions/packet-broker'

import PacketBroker from './admin-packet-broker'
import NetworkRoutingPolicy from './network-routing-policy'

const PacketBrokerRouter = ({ match }) => {
  useBreadcrumbs(
    'admin.packet-broker',
    <Breadcrumb path={'/admin/packet-broker'} content={sharedMessages.packetBroker} />,
  )

  return (
    <Require featureCheck={mayConfigurePacketBroker} otherwise={{ redirect: '/' }}>
      <RequireRequest requestAction={getPacketBrokerInfo()}>
        <Breadcrumbs />
        <Switch>
          <Route
            path={`${match.path}/networks/:netId/:tenantId?`}
            component={NetworkRoutingPolicy}
          />
          <Route path={`${match.path}`} component={PacketBroker} />
          <NotFoundRoute />
        </Switch>
      </RequireRequest>
    </Require>
  )
}

PacketBrokerRouter.propTypes = {
  match: PropTypes.match.isRequired,
}

export default PacketBrokerRouter
