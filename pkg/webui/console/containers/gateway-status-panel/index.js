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
import ReactApexChart from 'react-apexcharts'

import uptimeGraph from '@assets/misc/blurry-uptime.png'

import Panel from '@ttn-lw/components/panel'
import Icon, {
  IconDownlink,
  IconGateway,
  IconUplink,
  IconHeartRateMonitor,
  IconX,
  IconInfoCircle,
  IconBolt,
  IconRouterOff,
} from '@ttn-lw/components/icon'
import Tooltip from '@ttn-lw/components/tooltip'
import Button from '@ttn-lw/components/button'
import Status from '@ttn-lw/components/status'
import Spinner from '@ttn-lw/components/spinner'

import Message from '@ttn-lw/lib/components/message'
import DateTime from '@ttn-lw/lib/components/date-time'

import { isTranslated } from '@ttn-lw/lib/errors/utils'
import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

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
  frequencyRange: '{minFreq} - {maxFreq}MHz',
  protocol: 'Protocol {protocol}',
  percentage: '{percentage, number, percent}%',
  roundTripTimes: 'Roundtrip times (ms)',
  roundTripTimesTooltip:
    'The roundtrip times express the latency between the gateway and The Things Stack. The displayed value is a longest duration observed in the last 20 downlink operations. The round trip times can give you insight into how well and steady a connection to the gateway is.',
  transmissions: 'Transmissions',
  transmissionsTooltip:
    'In this section you can find information about the uplink and downlink count since the last reconnect of the gateway. The downlink count also tracks the number of acknowledged downlinks. It also contains information about when the last gateway status has been received, which is sent periodically depending on your gateway type and configuration.',
  noUplinks: 'No uplinks yet',
  noDownlinks: 'No downlinks yet',
  noStatus: 'No status',
  dutyCycleUtilization: 'Duty cycle utilization',
  dutyCycleUtilizationTooltip:
    'In this section you can find the duty cycle utilization of this gateway per sub-band. All network traffic has to be in accordance with local regulations that govern the maximum usage of radio transmissions per frequency in a given time-frame. The listing allows you to inspect how much of this allowance has been exhausted already. Once utilization is exhausted, you are required by law to cease transmissions by this gateway.',
  uptime: '30 day uptime',
  uptimeTooltip: 'The uptime of the gateway in the last 30 days.',
})

const options = {
  chart: {
    type: 'radialBar',
  },
  grid: {
    padding: {
      left: -9,
      right: -9,
      bottom: -12,
      top: -9,
    },
  },
  colors: [
    ({ value }) => {
      if (value < 55) {
        return '#1CB041'
      } else if (value === 100) {
        return '#DB2328'
      }

      return '#DB7600'
    },
  ],
  stroke: {
    lineCap: 'round',
  },
  dataLabels: {
    enabled: false,
  },
  legend: {
    show: false,
  },
  plotOptions: {
    radialBar: {
      track: {
        show: true,
        margin: 1.5,
      },
      dataLabels: {
        show: false,
      },
    },
  },
}

const SectionTitle = ({ title, tooltip }) => (
  <div>
    <Message content={title} className="fw-bold" />
    <Tooltip content={<Message content={tooltip} />}>
      <Icon icon={IconInfoCircle} className={style.gtwStatusPanelTooltip} />
    </Tooltip>
  </div>
)

SectionTitle.propTypes = {
  title: PropTypes.message.isRequired,
  tooltip: PropTypes.message.isRequired,
}

const GatewayStatusPanel = () => {
  const dispatch = useDispatch()
  const gatewayStats = useSelector(selectGatewayStatistics)
  const error = useSelector(selectGatewayStatisticsError)
  const fetching = useSelector(selectGatewayStatisticsIsFetching)
  const gtwId = useSelector(selectSelectedGatewayId)
  const isDisconnected = Boolean(gatewayStats?.disconnected_at)
  const isFetching = !Boolean(gatewayStats) && fetching
  const isUnavailable = Boolean(error) && Boolean(error.message) && isTranslated(error.message)

  /*   const connectedDays = useMemo(() => {
    if (gatewayStats?.connected_at) {
      const connectedDate = new Date(gatewayStats.connected_at)
      const currentDate = new Date()
      const diffTime = Math.abs(currentDate - connectedDate)
      return Math.ceil(diffTime / (1000 * 60 * 60 * 24))
    }
    return 0
  }, [gatewayStats]) */

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
    return gatewayStats && isNaN((mediumPoint * 100) / barWidth)
      ? 0
      : (mediumPoint * 100) / barWidth
  }, [gatewayStats, maxRoundTripTime, medianRoundTripTime, minRoundTripTime])

  const greenPointer = position <= 33
  const yellowPointer = position > 33 && position <= 66
  const redPointer = position > 66

  const showRoundTripTimes = Boolean(gatewayStats?.round_trip_times)
  const showUplink =
    Boolean(gatewayStats?.uplink_count) ||
    (typeof gatewayStats?.uplink_count === 'string' && gatewayStats?.uplink_count !== 0)
  const showUplinkTime = Boolean(gatewayStats?.last_uplink_received_at)
  const showDownlink =
    Boolean(gatewayStats?.downlink_count) ||
    (typeof gatewayStats?.downlink_count === 'string' && gatewayStats?.downlink_count !== 0)
  const showDownlinkTime = Boolean(gatewayStats?.last_downlink_received_at)
  const showStatus = Boolean(gatewayStats?.last_status_received_at)
  const showProtocol = Boolean(gatewayStats?.protocol)
  const showDutyCycleUtilization = Boolean(gatewayStats?.sub_bands)

  useEffect(() => {
    dispatch(startGatewayStatistics(gtwId))
    return () => {
      dispatch(stopGatewayStatistics())
    }
  }, [dispatch, gtwId])

  return (
    <Panel
      title="Gateway status"
      icon={isDisconnected ? IconRouterOff : IconGateway}
      shortCutLinkTitle="Network Operation Center"
      shortCutLinkPath="https://www.thethingsindustries.com/stack/features/noc"
      shortCutLinkTarget="_blank"
      divider
      className={style.gtwStatusPanel}
      iconClassName={isDisconnected ? style.gtwStatusPanelIcon : undefined}
      messageDecorators={
        <Status
          status={isDisconnected || isUnavailable ? 'bad' : isFetching ? 'mediocre' : 'green'}
          pulse
          big
          pulseTrigger={isDisconnected ? gatewayStats?.disconnected_at : gatewayStats?.connected_at}
        />
      }
    >
      {isFetching ? (
        <Spinner center inline>
          <Message content={sharedMessages.fetching} />
        </Spinner>
      ) : (
        <div className="grid">
          <div className="item-5">
            <SectionTitle title={m.uptime} tooltip={m.uptimeTooltip} />
            <img
              src={uptimeGraph}
              alt="Uptime graph"
              className={style.gtwStatusPanelUnlockUptimeImg}
            />
            <Message
              content={'Unlock uptime graph'}
              className={style.gtwStatusPanelUnlockUptimeMessage}
              component="div"
            />
            <Button.AnchorLink
              secondary
              message={'Upgrade now'}
              icon={IconBolt}
              href="https://www.thethingsindustries.com/stack/plans/"
              target="_blank"
              className="mt-cs-xs"
            />
          </div>
          <div className="item-5 item-start-8 d-flex direction-column j-between">
            <SectionTitle title={m.roundTripTimes} tooltip={m.roundTripTimesTooltip} />
            {showRoundTripTimes ? (
              <>
                <div>
                  <div className={style.gtwStatusPanelRoundTripTimeBar}>
                    <span
                      className={classNames(style.gtwStatusPanelRoundTripTimeBarPointer, {
                        'c-bg-success-normal': greenPointer,
                        'c-bg-warning-normal': yellowPointer,
                        'c-bg-error-normal': redPointer,
                      })}
                      style={{
                        left: `${position}%`,
                      }}
                    />
                  </div>
                  <div className="pos-relative d-flex j-between">
                    <span className="fs-s fw-bold">
                      <FormattedNumber value={(minRoundTripTime * 1000).toFixed(2)} />
                    </span>
                    <span
                      className={style.gtwStatusPanelRoundTripTimeBarMedian}
                      style={{
                        left: `${position < 21 ? 14 : position > 78 ? 71 : position - 7}%`,
                      }}
                    >
                      <FormattedNumber value={(medianRoundTripTime * 1000).toFixed(2)} />
                    </span>
                    <span className="fs-s fw-bold">
                      <FormattedNumber value={(maxRoundTripTime * 1000).toFixed(2)} />
                    </span>
                  </div>
                </div>
                <div
                  className={classNames(style.gtwStatusPanelRoundTripTimeTag, {
                    'c-text-success-normal': greenPointer,
                    'c-text-warning-normal': yellowPointer,
                    'c-text-error-normal': redPointer,
                  })}
                >
                  <FormattedNumber value={(medianRoundTripTime * 1000).toFixed(2)} />
                  <Message content="ms" />
                </div>
              </>
            ) : (
              <div>
                <Message content={'No data available'} className="fw-bold" component="div" />
                <Message
                  content={
                    'This gateway doesn’t have recent downlinks and cannot display the roundtrip time.'
                  }
                  className="fs-s c-text-neutral-light mt-cs-xs"
                  component="div"
                />
              </div>
            )}
          </div>
        </div>
      )}
      <hr className={style.gtwStatusPanelDivider} />
      {isFetching ? (
        <Spinner center>
          <Message content={sharedMessages.fetching} />
        </Spinner>
      ) : (
        <div className="grid">
          <div className="item-5">
            <div>
              <SectionTitle title={m.transmissions} tooltip={m.transmissionsTooltip} />
              <div className="d-flex al-center j-between gap-cs-m mb-cs-m mt-cs-l">
                <div className="d-flex al-center gap-cs-xxs">
                  <Icon icon={IconUplink} className="c-text-neutral-semilight" />
                  {!showUplink ? (
                    <Message content={m.noUplinks} className="fw-bold" />
                  ) : (
                    <FormattedNumber value={gatewayStats.uplink_count}>
                      {parts => (
                        <>
                          <b>{parts}</b>
                        </>
                      )}
                    </FormattedNumber>
                  )}
                </div>
                {showUplinkTime && (
                  <DateTime.Relative
                    value={gatewayStats.last_uplink_received_at}
                    relativeTimeStyle="short"
                  />
                )}
              </div>
              <div className="d-flex al-center j-between mb-cs-m gap-cs-m">
                <div className="d-flex al-center gap-cs-xxs">
                  <Icon icon={IconDownlink} className="c-text-neutral-semilight" />
                  {!showDownlink ? (
                    <Message content={m.noDownlinks} className="fw-bold" />
                  ) : (
                    <>
                      <FormattedNumber value={gatewayStats.downlink_count}>
                        {parts => (
                          <>
                            <b>{parts}</b>
                          </>
                        )}
                      </FormattedNumber>
                      {gatewayStats?.tx_acknowledgment_count && (
                        <>
                          {'('}
                          <FormattedNumber value={gatewayStats.tx_acknowledgment_count} />
                          {' Ack’d)'}
                        </>
                      )}
                    </>
                  )}
                </div>
                {showDownlinkTime && (
                  <DateTime.Relative
                    value={gatewayStats.last_downlink_received_at}
                    relativeTimeStyle="short"
                  />
                )}
              </div>
              <div className="d-flex al-center j-between gap-cs-m">
                <div className="d-flex al-center gap-cs-xxs fw-bold">
                  {showStatus ? (
                    <>
                      <Icon icon={IconHeartRateMonitor} className="c-text-success-normal" />
                      <Message content={m.statusRecieved} className="c-text-success-normal" />
                    </>
                  ) : isDisconnected ? (
                    <>
                      <Icon icon={IconX} className="c-text-error-normal" />
                      <Message content={'Disconnnected'} className="c-text-error-normal" />
                    </>
                  ) : (
                    <>
                      <Icon icon={IconX} className="c-text-error-normal" />
                      <Message content={m.noStatus} className="c-text-error-normal" />
                    </>
                  )}
                </div>
                {showStatus && (
                  <DateTime.Relative
                    value={gatewayStats.last_status_received_at}
                    relativeTimeStyle="short"
                  />
                )}
              </div>
            </div>
            <div className={style.gtwStatusPanelTag}>
              {showProtocol && (
                <Message
                  content={m.protocol}
                  values={{ protocol: gatewayStats.protocol.toUpperCase() }}
                  component="div"
                />
              )}
            </div>
          </div>
          <div className="item-5 item-start-8">
            <SectionTitle title={m.dutyCycleUtilization} tooltip={m.dutyCycleUtilizationTooltip} />
            {showDutyCycleUtilization ? (
              gatewayStats.sub_bands.map((band, index) => {
                const maxFrequency = band.max_frequency / 1e6
                const minFrequency = band.min_frequency / 1e6
                const utilization = band.downlink_utilization
                  ? (band.downlink_utilization * 100) / band.downlink_utilization_limit
                  : 0
                return (
                  <div
                    key={index}
                    className={classNames('d-flex al-center j-between fs-s', {
                      'mb-cs-m': index !== gatewayStats.sub_bands.length - 1,
                      'mt-cs-l': index === 0,
                    })}
                  >
                    <Message
                      content={m.frequencyRange}
                      values={{
                        minFreq: minFrequency.toFixed(1),
                        maxFreq: maxFrequency.toFixed(1),
                      }}
                      className="fs-s"
                    />
                    <div className="d-flex al-center j-center gap-cs-xs">
                      <ReactApexChart
                        options={options}
                        series={[utilization.toFixed(2)]}
                        type="radialBar"
                        height={20}
                        width={20}
                      />
                      <span
                        className={classNames('fs-s fw-bold', {
                          'c-text-success-normal': utilization <= 60,
                          'c-text-warning-normal': utilization > 60 && utilization < 100,
                          'c-text-error-normal': utilization === 100,
                        })}
                        style={{ minWidth: '39px' }}
                      >
                        <FormattedNumber
                          style="percent"
                          value={
                            isNaN(band.downlink_utilization / band.downlink_utilization_limit)
                              ? 0
                              : band.downlink_utilization / band.downlink_utilization_limit
                          }
                          minimumFractionDigits={2}
                        />
                      </span>
                    </div>
                  </div>
                )
              })
            ) : (
              <div>
                <Message
                  content={'No data available'}
                  className="fw-bold mt-cs-l"
                  component="div"
                />
                <Message
                  content={
                    'This gateway doesn’t have recent downlinks and cannot display the duty cycle utilization.'
                  }
                  className="fs-s c-text-neutral-light mt-cs-xs"
                  component="div"
                />
              </div>
            )}
          </div>
        </div>
      )}
    </Panel>
  )
}

export default GatewayStatusPanel
