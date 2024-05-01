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
import classNames from 'classnames'

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
  frequencyRange: '`{minFreq} - {maxFreq}MHz`',
})

const GatewayStatusPanel = () => {
  const dispatch = useDispatch()
  const gatewayStats = useSelector(selectGatewayStatistics)
  const error = useSelector(selectGatewayStatisticsError)
  const fetching = useSelector(selectGatewayStatisticsIsFetching)
  const gtwId = useSelector(selectSelectedGatewayId)
  const isDisconnected = !Boolean(gatewayStats) || Boolean(gatewayStats.disconnected_at)
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
  const maxRoundTripTime = useMemo(
    () =>
      gatewayStats?.round_trip_times && parseFloat(gatewayStats.round_trip_times.max.split('s')[0]),
    [gatewayStats],
  )
  const minRoundTripTime = useMemo(
    () =>
      gatewayStats?.round_trip_times && parseFloat(gatewayStats.round_trip_times.min.split('s')[0]),
    [gatewayStats],
  )
  const medianRoundTripTime = useMemo(
    () =>
      gatewayStats?.round_trip_times &&
      parseFloat(gatewayStats.round_trip_times.median.split('s')[0]),
    [gatewayStats],
  )
  const position = useMemo(() => {
    const barWidth = maxRoundTripTime - minRoundTripTime
    const mediumPoint = medianRoundTripTime - minRoundTripTime
    return gatewayStats && (mediumPoint * 100) / barWidth
  }, [gatewayStats, maxRoundTripTime, medianRoundTripTime, minRoundTripTime])

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
      <div className="d-flex">
        <div>
          <div>Uptime</div>
          <Message content="Transmissions" component="h4" className="mb-cs-xxs" />
          <div className="d-flex al-center j-between">
            <div className="d-flex al-center gap-cs-xxs">
              <Icon icon={IconUplink} />
              <FormattedNumber value={gatewayStats?.uplink_count ?? 0} />
            </div>
            {gatewayStats?.last_uplink_received_at && (
              <DateTime.Relative
                value={gatewayStats?.last_uplink_received_at}
                relativeTimeStyle="short"
              />
            )}
          </div>
          <div className="d-flex al-center j-between">
            <div className="d-flex al-center gap-cs-xxs">
              <Icon icon={IconDownlink} />
              <FormattedNumber value={gatewayStats?.downlink_count ?? 0} />
              {gatewayStats?.tx_acknowledgment_count && (
                <>
                  {'('}
                  <FormattedNumber value={gatewayStats?.tx_acknowledgment_count} />
                  {' Ack’d)'}
                </>
              )}
            </div>
            {gatewayStats?.last_downlink_received_at && (
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
              <Message content={'Connection'} component="h4" className="mb-cs-xxs" />
              {!isDisconnected && !isFetching && !isUnavailable && (
                <Message
                  content={'Connected to network'}
                  className={classNames(style.gtwStatusPanelTag, 'c-bg-success-normal')}
                />
              )}
            </div>
            <div>
              <Message content={'Protocol'} component="h4" className="mb-cs-xxs" />
              {gatewayStats?.protocol && (
                <Message
                  content={gatewayStats?.protocol.toUpperCase()}
                  className={classNames(style.gtwStatusPanelTag, 'c-bg-brand-normal')}
                />
              )}
            </div>
          </div>
          {Boolean(gatewayStats?.round_trip_times) && (
            <>
              <div className="d-flex j-between al-center">
                <Message
                  content="Round trip times (ms)"
                  component="h4"
                  className="mb-cs-xxs mt-0"
                />
                <Message
                  content={`(n=${gatewayStats?.round_trip_times.count})`}
                  className="mb-cs-xxs mt-0"
                />
              </div>
              <div className={style.gtwStatusPanelRoundTripTimeBar}>
                <span
                  className={style.gtwStatusPanelRoundTripTimeBarPointer}
                  style={{
                    left: `${position}%`,
                  }}
                />
              </div>
              <div className="pos-relative d-flex j-between">
                <span className="fs-s">
                  <FormattedNumber value={(minRoundTripTime * 1000).toFixed(2)} />
                </span>
                <span
                  className="pos-absolute fs-s fw-bold"
                  style={{
                    left: `${position - 6}%`,
                  }}
                >
                  <FormattedNumber value={(medianRoundTripTime * 1000).toFixed(2)} />
                </span>
                <span className="fs-s">
                  <FormattedNumber value={(maxRoundTripTime * 1000).toFixed(2)} />
                </span>
              </div>
            </>
          )}
          {Boolean(gatewayStats?.sub_bands) && (
            <div>
              <Message content={'Duty cycle utilization'} component="h4" className="mb-cs-xxs" />
              {gatewayStats.sub_bands.map(band => {
                const maxFrequency = band.max_frequency / 1e6
                const minFrequency = band.min_frequency / 1e6
                const utilization = band.downlink_utilization
                  ? (band.downlink_utilization * 100) / band.downlink_utilization_limit
                  : 0
                return (
                  <div key={band.id} className="d-flex al-center j-between">
                    <Message
                      content={m.frequencyRange}
                      values={{
                        minFreq: minFrequency.toFixed(1),
                        maxFreq: maxFrequency.toFixed(1),
                      }}
                      convertBackticks
                      className={style.gtwStatusPanelSubBand}
                    />
                    <div>{utilization.toFixed(2)}%</div>
                  </div>
                )
              })}
            </div>
          )}
        </div>
      </div>
    </Panel>
  )
}

export default GatewayStatusPanel
