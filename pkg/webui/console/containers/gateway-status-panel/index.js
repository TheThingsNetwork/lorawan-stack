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

import React, { useEffect, useMemo } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import { FormattedNumber, defineMessages } from 'react-intl'

import Panel from '@ttn-lw/components/panel'
import Icon, {
  IconDownlink,
  IconGateway,
  IconUplink,
  IconHeartRateMonitor,
} from '@ttn-lw/components/icon'
import StatusLabel from '@ttn-lw/components/status-label'

import Message from '@ttn-lw/lib/components/message'
import DateTime from '@ttn-lw/lib/components/date-time'

import { isTranslated } from '@ttn-lw/lib/errors/utils'

import { startGatewayStatistics, stopGatewayStatistics } from '@console/store/actions/gateways'

import {
  selectGatewayStatistics,
  selectGatewayStatisticsError,
  selectGatewayStatisticsIsFetching,
  selectSelectedGatewayId,
} from '@console/store/selectors/gateways'

import style from './gateway-status-panel.styl'

const m = defineMessages({
  statusLabel: 'Up and running for {days} days',
  statusRecieved: 'Status received',
})

const GatewayStatusPanel = () => {
  const gatewayStats = useSelector(selectGatewayStatistics)
  const error = useSelector(selectGatewayStatisticsError)
  const fetching = useSelector(selectGatewayStatisticsIsFetching)
  const gtwId = useSelector(selectSelectedGatewayId)
  const isDisconnected = Boolean(gatewayStats) && Boolean(gatewayStats.disconnected_at)
  const isFetching = !Boolean(gatewayStats) && fetching
  const isUnavailable = Boolean(error) && Boolean(error.message) && isTranslated(error.message)

  const connectedDays = useMemo(() => {
    if (gatewayStats?.connected_at) {
      const connectedDate = new Date(gatewayStats.connected_at)
      const currentDate = new Date()
      const diffTime = Math.abs(currentDate - connectedDate)
      return Math.ceil(diffTime / (1000 * 60 * 60 * 24))
    }
    return 0
  }, [gatewayStats])

  const dispatch = useDispatch()
  useEffect(() => {
    dispatch(startGatewayStatistics(gtwId))
    return () => {
      dispatch(stopGatewayStatistics())
    }
  }, [dispatch, gtwId])

  return (
    <Panel
      title="Gateway Status"
      shortCutLinkPath="/"
      icon={IconGateway}
      shortCutLinkTitle="Network Operation Center"
      target="_blank"
    >
      <StatusLabel success content={m.statusLabel} contentValues={{ days: connectedDays }} />
      <hr className={style.gtwStatusPanelDivider} />
      <div>
        <div>
          <div></div>
          <Message content="Transmissions" component="h4" className="mb-cs-xs" />
          <div className="d-flex al-center j-between">
            <div className="d-flex al-center gap-cs-xxs">
              <Icon icon={IconUplink} />
              <FormattedNumber value={gatewayStats?.uplink_count} />
            </div>
            {gatewayStats?.last_status_received_at && (
              <DateTime.Relative
                value={gatewayStats?.last_uplink_received_at}
                relativeTimeStyle="short"
              />
            )}
          </div>
          <div className="d-flex al-center j-between">
            <div className="d-flex al-center gap-cs-xxs">
              <Icon icon={IconDownlink} />
              <FormattedNumber value={gatewayStats?.downlink_count} />
              {gatewayStats?.tx_acknowledgment_count && (
                <>
                  {'('}
                  <FormattedNumber value={gatewayStats?.tx_acknowledgment_count} />
                  {' Ack’d)'}
                </>
              )}
            </div>
            {gatewayStats?.last_status_received_at && (
              <DateTime.Relative
                value={gatewayStats?.last_downlink_received_at}
                relativeTimeStyle="short"
              />
            )}
          </div>
          <div className="d-flex al-center j-between">
            <div className="d-flex al-center gap-cs-xxs">
              <Icon icon={IconHeartRateMonitor} />
              <Message content={m.statusRecieved} />
            </div>
            {gatewayStats?.last_status_received_at && (
              <DateTime.Relative
                value={gatewayStats?.last_status_received_at}
                relativeTimeStyle="short"
              />
            )}
          </div>
        </div>
        <div>
          <div className="d-flex al-center">
            <div>
              <Message content={'Connection'} component="h4" className="mb-cs-xs" />
              {isDisconnected && <Message content={'Disconnected'} className="c-bg-error-normal" />}
              {isFetching && <Message content={'Fetching'} className="c-bg-warning-normal" />}
              {isUnavailable && <Message content={'Unavailable'} />}
              {!isDisconnected && !isFetching && !isUnavailable && (
                <Message content={'Connected to network'} className="c-bg-success-normal " />
              )}
            </div>
            <div>
              <Message content={'Protocol'} />
            </div>
          </div>
          <div></div>
          <div></div>
        </div>
      </div>
    </Panel>
  )
}

export default GatewayStatusPanel
