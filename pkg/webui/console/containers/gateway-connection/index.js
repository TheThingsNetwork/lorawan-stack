// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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
import { useDispatch, useSelector } from 'react-redux'

import Status from '@ttn-lw/components/status'
import Icon from '@ttn-lw/components/icon'
import DocTooltip from '@ttn-lw/components/tooltip/doc'
import Tooltip from '@ttn-lw/components/tooltip'

import Message from '@ttn-lw/lib/components/message'

import LastSeen from '@console/components/last-seen'

import useConnectionReactor from '@console/containers/gateway-connection/use-connection-reactor'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import { isNotFoundError, isTranslated } from '@ttn-lw/lib/errors/utils'
import { selectGsConfig } from '@ttn-lw/lib/selectors/env'
import getHostFromUrl from '@ttn-lw/lib/host-from-url'

import { startGatewayStatistics, stopGatewayStatistics } from '@console/store/actions/gateways'

import {
  selectGatewayById,
  selectGatewayStatistics,
  selectGatewayStatisticsError,
  selectGatewayStatisticsIsFetching,
} from '@console/store/selectors/gateways'
import { selectGatewayLastSeen } from '@console/store/selectors/gateway-status'

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
  const { className, gtwId } = props

  const gateway = useSelector(state => selectGatewayById(state, gtwId))
  const gsConfig = selectGsConfig()
  const consoleGsAddress = getHostFromUrl(gsConfig.base_url)
  const gatewayServerAddress = getHostFromUrl(gateway.gateway_server_address)
  const statistics = useSelector(selectGatewayStatistics)
  const error = useSelector(selectGatewayStatisticsError)
  const fetching = useSelector(selectGatewayStatisticsIsFetching)
  const lastSeen = useSelector(selectGatewayLastSeen)
  const isOtherCluster = consoleGsAddress !== gatewayServerAddress

  const dispatch = useDispatch()

  useConnectionReactor(gtwId)

  useEffect(() => {
    dispatch(startGatewayStatistics(gtwId))
    return () => {
      dispatch(stopGatewayStatistics())
    }
  }, [dispatch, gtwId])

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
  gtwId: PropTypes.string.isRequired,
}

GatewayConnection.defaultProps = {
  className: undefined,
}

export default GatewayConnection
