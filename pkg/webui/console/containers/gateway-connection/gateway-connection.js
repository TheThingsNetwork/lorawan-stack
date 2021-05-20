// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
import classnames from 'classnames'
import { FormattedNumber } from 'react-intl'

import Status from '@ttn-lw/components/status'
import Icon from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'

import LastSeen from '@console/components/last-seen'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import { isNotFoundError, isTranslated } from '@ttn-lw/lib/errors/utils'

import style from './gateway-connection.styl'

class GatewayConnection extends React.PureComponent {
  static propTypes = {
    className: PropTypes.string,
    error: PropTypes.oneOfType([PropTypes.error, PropTypes.shape({ message: PropTypes.message })]),
    fetching: PropTypes.bool,
    lastSeen: PropTypes.instanceOf(Date),
    startStatistics: PropTypes.func.isRequired,
    statistics: PropTypes.gatewayStats,
    stopStatistics: PropTypes.func.isRequired,
  }

  static defaultProps = {
    className: undefined,
    fetching: false,
    error: null,
    statistics: null,
    lastSeen: undefined,
  }

  componentDidMount() {
    const { startStatistics } = this.props

    startStatistics()
  }

  componentWillUnmount() {
    const { stopStatistics } = this.props

    stopStatistics()
  }

  get status() {
    const { statistics, error, fetching, lastSeen } = this.props

    const isNotConnected = Boolean(error) && isNotFoundError(error)
    const isFetching = !Boolean(statistics) && fetching
    const isUnavailable = Boolean(error) && Boolean(error.message) && isTranslated(error.message)
    const hasStatistics = Boolean(statistics)
    const hasLastSeen = Boolean(lastSeen)

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
      message = sharedMessages.connected
      statusIndicator = 'good'
    } else {
      message = sharedMessages.unknown
      statusIndicator = 'unknown'
    }

    return (
      <Status className={style.status} status={statusIndicator} flipped>
        {statusIndicator === 'good' && hasLastSeen ? (
          <LastSeen lastSeen={lastSeen} />
        ) : (
          <Message content={message} />
        )}
      </Status>
    )
  }

  get messages() {
    const { statistics } = this.props

    if (!statistics) {
      return null
    }

    const uplinks = statistics.uplink_count || '0'
    const downlinks = statistics.downlink_count || '0'

    const uplinkCount = parseInt(uplinks) || 0
    const downlinkCount = parseInt(downlinks) || 0

    return (
      <div className={style.messages}>
        <span className={style.messageCount}>
          <Icon className={style.icon} icon="uplink" />
          <FormattedNumber value={uplinkCount} />
        </span>
        <span className={style.messageCount}>
          <Icon className={style.icon} icon="downlink" />
          <FormattedNumber value={downlinkCount} />
        </span>
      </div>
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

export default GatewayConnection
