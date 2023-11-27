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

import { connect } from 'react-redux'

import { selectGsConfig } from '@ttn-lw/lib/selectors/env'
import getHostFromUrl from '@ttn-lw/lib/host-from-url'

import {
  startGatewayStatistics,
  stopGatewayStatistics,
  updateGatewayStatistics,
} from '@console/store/actions/gateways'

import {
  selectGatewayStatistics,
  selectGatewayStatisticsError,
  selectGatewayStatisticsIsFetching,
  selectLatestGatewayEvent,
  selectGatewayById,
} from '@console/store/selectors/gateways'
import { selectGatewayLastSeen } from '@console/store/selectors/gateway-status'

export default GatewayConnection =>
  connect(
    (state, ownProps) => {
      const gateway = selectGatewayById(state, ownProps.gtwId)
      const gsConfig = selectGsConfig()
      const consoleGsAddress = getHostFromUrl(gsConfig.base_url)
      const gatewayServerAddress = getHostFromUrl(gateway.gateway_server_address)

      return {
        statistics: selectGatewayStatistics(state, ownProps),
        error: selectGatewayStatisticsError(state, ownProps),
        fetching: selectGatewayStatisticsIsFetching(state, ownProps),
        latestEvent: selectLatestGatewayEvent(state, ownProps.gtwId),
        lastSeen: selectGatewayLastSeen(state),
        isOtherCluster: consoleGsAddress !== gatewayServerAddress,
      }
    },
    (dispatch, ownProps) => ({
      startStatistics: () => dispatch(startGatewayStatistics(ownProps.gtwId)),
      stopStatistics: () => dispatch(stopGatewayStatistics()),
      updateGatewayStatistics: () => dispatch(updateGatewayStatistics(ownProps.gtwId)),
    }),
  )(GatewayConnection)
