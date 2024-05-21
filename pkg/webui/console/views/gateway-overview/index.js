// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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
import { useSelector } from 'react-redux'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'

import BlurryNetworkActivityPanel from '@console/components/blurry-network-activity-panel'
import GatewayMapPanel from '@console/components/gateway-map-panel'

import GatewayOverviewHeader from '@console/containers/gateway-overview-header'
import GatewayStatusPanel from '@console/containers/gateway-status-panel'
import GatewayGeneralInformationPanel from '@console/containers/gateway-general-information-panel'

import Require from '@console/lib/components/require'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { mayViewGatewayInfo } from '@console/lib/feature-checks'

import { selectSelectedGateway } from '@console/store/selectors/gateways'

const GatewayOverview = () => {
  const gateway = useSelector(selectSelectedGateway)

  return (
    <Require featureCheck={mayViewGatewayInfo} otherwise={{ redirect: '/' }}>
      <IntlHelmet title={sharedMessages.overview} />
      <GatewayOverviewHeader gateway={gateway} />
      <div className="container container--xl grid p-ls-s gap-ls-s">
        <div className="item-12 md:item-12 lg:item-6 sm:item-6">
          <GatewayGeneralInformationPanel />
        </div>
        <div className="item-12 md:item-12 lg:item-6 sm:item-6">
          <GatewayStatusPanel />
        </div>
        <div className="item-12 md:item-12 lg:item-6 sm:item-6">
          <BlurryNetworkActivityPanel />
        </div>
        <div className="item-12 md:item-12 lg:item-6 sm:item-6">
          <GatewayMapPanel gateway={gateway} />
        </div>
      </div>
    </Require>
  )
}

export default GatewayOverview
