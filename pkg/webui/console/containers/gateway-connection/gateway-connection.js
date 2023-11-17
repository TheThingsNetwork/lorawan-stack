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

import React, { useEffect, useMemo } from 'react'
import classnames from 'classnames'
import { FormattedNumber, defineMessages } from 'react-intl'

import Status from '@ttn-lw/components/status'
import Icon from '@ttn-lw/components/icon'
import DocTooltip from '@ttn-lw/components/tooltip/doc'
import Tooltip from '@ttn-lw/components/tooltip'

import Message from '@ttn-lw/lib/components/message'

import LastSeen from '@console/components/last-seen'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import { isNotFoundError, isTranslated } from '@ttn-lw/lib/errors/utils'

import style from './gateway-connection.styl'

const m = defineMessages({
  lastSeenAvailableTooltip:
    'The elapsed time since the network registered the last activity of this gateway. This is determined from received uplinks, or sent status messages of this gateway.',
  disconnectedTooltip:
    'The gateway has currently no TCP connection established with the Gateway Server. For (rare) UDP based gateways, this can also mean that the gateway initiated no pull/push data request within the last 30 seconds.',
  connectedTooltip:
    'This gateway is connected to the Gateway Server but the network has not registered any activity (sent uplinks or status messages) from it yet.',
  otherClusterTooltip:
    'This gateway is connected to an external Gateway Server that is not handling messages for this cluster. You will hence not be able to see any activity from this gateway.',
  messageCountTooltip:
    'The amount of received uplinks and sent downlinks of this gateway since the last (re)connect. Note that some gateway types reconnect frequently causing the counter to be reset.',
})

const GatewayConnection = props => {
  const {
    startStatistics,
    stopStatistics,
    statistics,
    error,
    fetching,
    lastSeen,
    isOtherCluster,
    className,
  } = props

  useEffect(() => {
    startStatistics()
    return () => {
      stopStatistics()
    }
  }, [startStatistics, stopStatistics])

  const status = useMemo(() => {
    const statsNotFound = Boolean(error) && isNotFoundError(error)
    const isDisconnected = Boolean(statistics) && Boolean(statistics.disconnected_at)
    const isFetching = !Boolean(statistics) && fetching
    const isUnavailable = Boolean(error) && Boolean(error.message) && isTranslated(error.message)
    const hasStatistics = Boolean(statistics)
    const hasLastSeen = Boolean(lastSeen)

    let statusIndicator = null
    let message = null
    let tooltipMessage = undefined
    let docPath = '/getting-started/console/troubleshooting'
    let docTitle = sharedMessages.troubleshooting

    if (statsNotFound) {
      statusIndicator = 'bad'
      message = sharedMessages.disconnected
      tooltipMessage = m.disconnectedTooltip
      docPath = '/gateways/troubleshooting/#my-gateway-wont-connect-what-do-i-do'
    } else if (isDisconnected) {
      tooltipMessage = m.disconnectedTooltip
      docPath = '/gateways/troubleshooting/#my-gateway-wont-connect-what-do-i-do'
    } else if (isFetching) {
      statusIndicator = 'mediocre'
      message = sharedMessages.connecting
    } else if (isUnavailable) {
      statusIndicator = 'unknown'
      message = error.message
      if (isOtherCluster) {
        tooltipMessage = m.otherClusterTooltip
        docPath = '/gateways/troubleshooting/#my-gateway-shows-a-other-cluster-status-why'
      }
    } else if (hasStatistics) {
      message = sharedMessages.connected
      statusIndicator = 'good'
      if (hasLastSeen) {
        tooltipMessage = m.lastSeenAvailableTooltip
      } else {
        docPath =
          'gateways/troubleshooting/#my-gateway-is-shown-as-connected-in-the-console-but-i-dont-see-any-events-including-the-gateway-connection-stats-what-do-i-do'
        tooltipMessage = m.connectedTooltip
      }
      docTitle = sharedMessages.moreInformation
    } else {
      message = sharedMessages.unknown
      statusIndicator = 'unknown'
      docPath = '/gateways/troubleshooting'
    }

    let node

    if (isDisconnected) {
      node = (
        <LastSeen
          status="bad"
          message={sharedMessages.disconnected}
          lastSeen={statistics.disconnected_at}
          flipped
        >
          <Icon icon="help_outline" textPaddedLeft small nudgeUp className="tc-subtle-gray" />
        </LastSeen>
      )
    } else if (statusIndicator === 'good' && hasLastSeen) {
      node = (
        <LastSeen lastSeen={lastSeen} flipped>
          <Icon icon="help_outline" textPaddedLeft small nudgeUp className="tc-subtle-gray" />
        </LastSeen>
      )
    } else {
      node = (
        <Status className={style.status} status={statusIndicator} label={message} flipped>
          <Icon icon="help_outline" textPaddedLeft small nudgeUp className="tc-subtle-gray" />
        </Status>
      )
    }

    if (tooltipMessage) {
      return (
        <DocTooltip
          docPath={docPath}
          docTitle={docTitle}
          content={<Message content={tooltipMessage} />}
          children={node}
        />
      )
    }

    return node
  }, [error, fetching, isOtherCluster, lastSeen, statistics])

  const messages = useMemo(() => {
    if (!statistics) {
      return null
    }

    const uplinks = statistics.uplink_count || '0'
    const downlinks = statistics.downlink_count || '0'

    const uplinkCount = parseInt(uplinks) || 0
    const downlinkCount = parseInt(downlinks) || 0

    return (
      <Tooltip content={<Message content={m.messageCountTooltip} />}>
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
      </Tooltip>
    )
  }, [statistics])

  return (
    <div className={classnames(className, style.container)}>
      {messages}
      {status}
    </div>
  )
}

GatewayConnection.propTypes = {
  className: PropTypes.string,
  error: PropTypes.oneOfType([PropTypes.error, PropTypes.shape({ message: PropTypes.message })]),
  fetching: PropTypes.bool,
  isOtherCluster: PropTypes.bool.isRequired,
  lastSeen: PropTypes.oneOfType([
    PropTypes.string,
    PropTypes.number, // Support timestamps.
    PropTypes.instanceOf(Date),
  ]),
  startStatistics: PropTypes.func.isRequired,
  statistics: PropTypes.gatewayStats,
  stopStatistics: PropTypes.func.isRequired,
}

GatewayConnection.defaultProps = {
  className: undefined,
  fetching: false,
  error: null,
  statistics: null,
  lastSeen: undefined,
}

export default GatewayConnection
