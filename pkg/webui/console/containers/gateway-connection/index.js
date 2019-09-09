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
import { connect } from 'react-redux'
import classnames from 'classnames'

import sharedMessages from '../../../lib/shared-messages'
import PropTypes from '../../../lib/prop-types'
import Status from '../../../components/status'
import Icon from '../../../components/icon'
import DateTime from '../../../lib/components/date-time'
import Message from '../../../lib/components/message'
import Button from '../../../components/button'
import { isNotFoundError, isTranslated } from '../../../lib/errors/utils'

import {
  selectGatewayStatistics,
  selectGatewayStatisticsError,
  selectGatewayStatisticsIsFetching,
} from '../../store/selectors/gateways'
import { startGatewayStatistics, stopGatewayStatistics } from '../../store/actions/gateways'

import style from './gateway-connection.styl'

class GatewayConnection extends React.PureComponent {
  componentDidMount() {
    const { startStatistics } = this.props

    startStatistics()
  }

  componentWillUnmount() {
    const { stopStatistics } = this.props

    stopStatistics()
  }

  get status() {
    const { statistics, error, fetching } = this.props

    const isNotConnected = Boolean(error) && isNotFoundError(error)
    const isFetching = !Boolean(statistics) && fetching
    const isUnavailable = Boolean(error) && Boolean(error.message) && isTranslated(error.message)
    const hasStatistics = Boolean(statistics)

    let statusIndicator = null
    let message = null

    if (isNotConnected) {
      statusIndicator = 'bad'
      message = sharedMessages.disconnected
    } else if (isFetching) {
      statusIndicator = 'mediocre'
      message = sharedMessages.connecting
    } else if (isUnavailable) {
      statusIndicator = 'unknown'
      message = error.message
    } else if (hasStatistics) {
      message = sharedMessages.lastSeen
      statusIndicator = 'good'
    } else {
      message = sharedMessages.unknown
      statusIndicator = 'unknown'
    }

    return (
      <Status className={style.status} status={statusIndicator}>
        <Message className={style.lastSeen} content={message} />
        {statusIndicator === 'good' && (
          <DateTime.Relative value={statistics.last_status_received_at} />
        )}
      </Status>
    )
  }

  get messages() {
    const { statistics, error, startStatistics, fetching } = this.props

    if (isNotFoundError(error)) {
      return <Button naked secondary disabled={fetching} icon="refresh" onClick={startStatistics} />
    }

    if (!statistics) {
      return null
    }

    const uplinkCount = statistics.uplink_count || 0
    const downlinkCount = statistics.downlink_count || 0

    return (
      <React.Fragment>
        <span className={style.messageCount}>
          <Icon className={style.icon} icon="uplink" />
          <span>{uplinkCount}</span>
        </span>
        <span className={style.messageCount}>
          <Icon className={style.icon} icon="downlink" />
          <span>{downlinkCount}</span>
        </span>
      </React.Fragment>
    )
  }

  render() {
    const { className } = this.props

    return (
      <div className={classnames(className, style.container)}>
        {this.status}
        {this.messages}
      </div>
    )
  }
}

GatewayConnection.propTypes = {
  className: PropTypes.string,
  error: PropTypes.error,
  fetching: PropTypes.bool,
  startStatistics: PropTypes.func.isRequired,
  statistics: PropTypes.gatewayStats,
  stopStatistics: PropTypes.func.isRequired,
}

GatewayConnection.defaultProps = {
  className: undefined,
  fetching: false,
  error: null,
  statistics: null,
}

export default connect(
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
