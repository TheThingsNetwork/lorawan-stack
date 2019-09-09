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

import {
  isGsStatusReceiveEvent,
  isGsDownlinkSendEvent,
  isGsUplinkReceiveEvent,
} from '../../../lib/selectors/event'
import PropTypes from '../../../lib/prop-types'

/**
 * `withConnectionReactor` is a HOC that handles gateway connection statistics updates based on
 * gateway uplink, downlink and connection events.
 * @param {Object} Component - React component to be wrapped by the reactor.
 * @returns {Object} - A wrapped react component.
 */
const withConnectionReactor = Component => {
  class ConnectionReactor extends React.PureComponent {
    componentDidUpdate(prevProps) {
      const { latestEvent, updateGatewayStatistics } = this.props

      if (Boolean(latestEvent) && latestEvent !== prevProps.latestEvent) {
        const { name } = latestEvent
        const isHeartBeatEvent =
          isGsDownlinkSendEvent(name) ||
          isGsUplinkReceiveEvent(name) ||
          isGsStatusReceiveEvent(name)

        if (isHeartBeatEvent) {
          updateGatewayStatistics()
        }
      }
    }

    render() {
      const { latestEvent, updateGatewayStatistics, ...rest } = this.props
      return <Component {...rest} />
    }
  }

  ConnectionReactor.propTypes = {
    latestEvent: PropTypes.event,
    updateGatewayStatistics: PropTypes.func.isRequired,
  }

  ConnectionReactor.defaultProps = {
    latestEvent: undefined,
  }

  return ConnectionReactor
}

export default withConnectionReactor
