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

import {
  selectGatewayStatistics,
  selectGatewayStatisticsError,
  selectGatewayStatisticsIsFetching,
} from '../../store/selectors/gateways'
import { startGatewayStatistics, stopGatewayStatistics } from '../../store/actions/gateways'

export default GatewayConnection =>
  connect(
    function(state, props) {
      return {
        statistics: selectGatewayStatistics(state, props),
        error: selectGatewayStatisticsError(state, props),
        fetching: selectGatewayStatisticsIsFetching(state, props),
      }
    },
    (dispatch, ownProps) => ({
      startStatistics: () => dispatch(startGatewayStatistics(ownProps.gtwId)),
      stopStatistics: () => dispatch(stopGatewayStatistics()),
    }),
  )(GatewayConnection)
