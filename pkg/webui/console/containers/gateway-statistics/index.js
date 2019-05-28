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
import bind from 'autobind-decorator'
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
  statisticsSelector,
  statisticsErrorSelector,
  statisticsIsAvailableSelector,
  statisticsIsFetchingSelector,
} from '../../store/selectors/gateway'
import {
  startGatewayStatistics,
  stopGatewayStatistics,
} from '../../store/actions/gateway'

import style from './gateway-statistics.styl'

@bind
class GatewayStatistic extends React.PureComponent {

  componentDidMount () {
    const { startStatistics } = this.props

    startStatistics()
  }

  componentWillUnmount () {
    const { stopStatistics } = this.props

    stopStatistics()
  }

  get status () {
    const { statistics, error, fetching } = this.props

    let statusIndicator = null
    let message = null

    if (isNotFoundError(error)) {
      statusIndicator = 'bad'
      message = sharedMessages.disconnected
    } else if (fetching) {
      statusIndicator = 'mediocre'
      message = sharedMessages.connecting
    } else if (error && error.message && isTranslated(error.message)) {
      statusIndicator = 'unknown'
      message = error.message
    } else if (statistics) {
      message = sharedMessages.lastSeen
      statusIndicator = 'good'
    } else {
      message = sharedMessages.unknown
      statusIndicator = 'unknown'
    }

    return (
      <Status
        className={style.status}
        status={statusIndicator}
      >
        <Message
          className={style.lastSeen}
          content={message}
        />
        { statusIndicator === 'good' && (
          <DateTime.Relative
            value={statistics.last_status_received_at}
          />
        )}
      </Status>
    )
  }

  get messages () {
    const {
      statistics,
      error,
      startStatistics,
      fetching,
    } = this.props

    if (isNotFoundError(error)) {
      return (
        <Button
          naked
          secondary
          disabled={fetching}
          icon="refresh"
          onClick={startStatistics}
        />
      )
    }

    if (!statistics) {
      return null
    }

    const uplinkCount = statistics.uplink_count
    const downlinkCount = statistics.downlink_count

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

  render () {
    const { className, available } = this.props

    if (!available) {
      return null
    }

    return (
      <div className={classnames(className, style.container)}>
        {this.status}
        {this.messages}
      </div>
    )
  }
}

GatewayStatistic.propTypes = {
  gtwId: PropTypes.string.isRequired,
  startStatistics: PropTypes.func.isRequired,
  stopStatistics: PropTypes.func.isRequired,
  fetching: PropTypes.bool,
  available: PropTypes.bool,
  error: PropTypes.object,
  statistics: PropTypes.object,
}

GatewayStatistic.defaultProps = {
  fetching: false,
  available: false,
  error: null,
  statistics: null,
}

export default connect(function (state, props) {
  return {
    statistics: statisticsSelector(state, props),
    error: statisticsErrorSelector(state, props),
    available: statisticsIsAvailableSelector(state, props),
    fetching: statisticsIsFetchingSelector(state, props),
  }
},
(dispatch, ownProps) => ({
  startStatistics: () => dispatch(startGatewayStatistics(ownProps.gtwId)),
  stopStatistics: () => dispatch(stopGatewayStatistics()),
}))(GatewayStatistic)
