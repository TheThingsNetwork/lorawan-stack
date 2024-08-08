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
import { defineMessages } from 'react-intl'
import { useParams } from 'react-router-dom'

import Panel from '@ttn-lw/components/panel'
import Icon, { IconGateway, IconInfoCircle, IconBolt, IconRouterOff } from '@ttn-lw/components/icon'
import Tooltip from '@ttn-lw/components/tooltip'
import Button from '@ttn-lw/components/button'
import Status from '@ttn-lw/components/status'
import Link from '@ttn-lw/components/link'
import Spinner from '@ttn-lw/components/spinner'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import { getBackendErrorName, isBackend } from '@ttn-lw/lib/errors/utils'

import { startGatewayStatistics, stopGatewayStatistics } from '@console/store/actions/gateways'

import {
  selectGatewayStatistics,
  selectGatewayStatisticsError,
  selectGatewayStatisticsIsFetching,
} from '@console/store/selectors/gateways'

import Transmissions from './transmissions'
import RoundtripTimes from './roundtrip-times'
import DutyCycleUtilization from './duty-cycle-utilization'

import style from './gateway-status-panel.styl'

const m = defineMessages({
  protocol: 'Protocol {protocol}',
  roundTripTimes: 'Roundtrip times (ms)',
  roundTripTimesTooltip:
    '<b>What is this?</b>{lineBreak}The roundtrip times express the latency between the gateway and The Things Stack.{lineBreak}The displayed values show the longest (P95 to account for outliers) and shortest roundtrip durations as well as the median observed in the last 20 downlink operations.{lineBreak}The roundtrip times can give you insight into how well and steady a connection to the gateway is.',
  transmissions: 'Connection stats',
  transmissionsTooltip:
    '<b>What is this?</b>{lineBreak}Information about the uplink and downlink count since the last reconnect of the gateway. The downlink count also tracks the number of confirmed downlinks. It also contains information about when the last gateway status has been received, which is sent periodically or once on first connection depending on your gateway type.',
  dutyCycleUtilization: 'Duty cycle utilization',
  dutyCycleUtilizationTooltip:
    '<b>What is this?</b>{lineBreak}The utilization of this gateway per sub-band.{lineBreak}All network traffic has to be in accordance with local regulations that govern the maximum usage of radio transmissions per frequency in a given time-frame. The listing allows you to inspect how much of this allowance has been exhausted already. Once utilization is exhausted, transmissions are suspended automatically by the gateway server.',
  uptime: '30 day uptime',
  uptimeTooltip:
    '<b>What is this?</b>{lineBreak}The 30 day uptime expresses the relative amount of time that the gateway has been connected to the gateway server in the last 30 days.',
  noRoundtrip: 'This gateway doesn’t have recent downlinks and cannot display the roundtrip time.',
  noDutyCycle:
    'This gateway does not have recent downlinks and cannot display the duty cycle utilization.',
  unlockGraph: 'Unlock uptime graph',
  noConnection: 'This gateway has not made any connection attempts yet.',
  noConnectionDescription:
    'If you have recently registered this gateway, please wait for a few moments to allow the gateway to connect. Otherwise please refer to our <Link>gateway troubleshooting documentation</Link>.',
})

const SectionTitle = ({ title, tooltip }) => (
  <div className={style.sectionTitle}>
    <Message content={title} className="fw-bold" />
    <Tooltip
      content={
        <Message
          content={tooltip}
          values={{
            lineBreak: <br />,
            b: chunks => <b>{chunks}</b>,
          }}
        />
      }
    >
      <Icon icon={IconInfoCircle} small className={style.gtwStatusPanelTooltip} />
    </Tooltip>
  </div>
)

SectionTitle.propTypes = {
  title: PropTypes.message.isRequired,
  tooltip: PropTypes.message.isRequired,
}

const EmptyState = ({ title, message }) => (
  <div className="mt-cs-l">
    <Message content={title} component="div" className="c-text-neutral-semilight" />
    <Message content={message} className="fs-s c-text-neutral-semilight mt-cs-s" component="div" />
  </div>
)

EmptyState.propTypes = {
  message: PropTypes.message.isRequired,
  title: PropTypes.message.isRequired,
}

const GatewayStatusPanel = () => {
  const dispatch = useDispatch()
  const { gtwId } = useParams()
  const gatewayStats = useSelector(selectGatewayStatistics)
  const error = useSelector(selectGatewayStatisticsError)
  const fetching = useSelector(selectGatewayStatisticsIsFetching)
  const isDisconnected = Boolean(gatewayStats?.disconnected_at)
  const isFetching = !Boolean(gatewayStats) && fetching
  const noConnectionYet = useMemo(
    () => isBackend(error) && getBackendErrorName(error).includes('not_connected'),
    [error],
  )

  const isUnavailable = Boolean(error) && Boolean(error.message)

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

  const showRoundTripTimes = Boolean(gatewayStats?.round_trip_times)
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
      shortCutLinkTitle={sharedMessages.noc}
      shortCutLinkPath="https://www.thethingsindustries.com/stack/features/noc"
      shortCutLinkTarget="_blank"
      divider
      className={style.gtwStatusPanel}
      iconClassName={isDisconnected ? style.gtwStatusPanelIcon : undefined}
      messageDecorators={
        <Status
          status={
            isDisconnected
              ? 'bad'
              : isFetching || isUnavailable || noConnectionYet
                ? 'mediocre'
                : 'green'
          }
          pulse
          big
          pulseTrigger={gatewayStats?.last_status_received_at}
        />
      }
    >
      {isFetching ? (
        <Spinner center inline>
          <Message content={sharedMessages.fetching} />
        </Spinner>
      ) : noConnectionYet ? (
        <div className="d-flex j-center al-center flex-grow">
          <div className="d-flex direction-column j-center al-center text-center w-60 gap-cs-m mb-ls-m">
            <Message content={m.noConnection} className="fw-bold fs-l lh-xs3" component="div" />
            <Message
              content={m.noConnectionDescription}
              className="c-text-neutral-light lh-xxs"
              values={{
                Link: msg => (
                  <Link.DocLink primary path="/gateways/troubleshooting/" allowWrap>
                    {msg}
                  </Link.DocLink>
                ),
              }}
            />
          </div>
        </div>
      ) : (
        <>
          <div className={style.gtwStatusPanelUpperContainer}>
            <div className="d-flex direction-column j-between w-full sm-md:j-start">
              <SectionTitle title={m.uptime} tooltip={m.uptimeTooltip} />
              <div className="mt-cs-l">
                <Message content={m.unlockGraph} component="div" className="fw-bold" />
                <Button.AnchorLink
                  naked
                  message={sharedMessages.upgradeNow}
                  icon={IconBolt}
                  href="https://www.thethingsindustries.com/stack/plans/"
                  target="_blank"
                  className={style.gtwStatusPanelUpgradeButton}
                />
              </div>
            </div>
            <div className="d-flex direction-column j-between w-full">
              <SectionTitle title={m.roundTripTimes} tooltip={m.roundTripTimesTooltip} />
              {showRoundTripTimes ? (
                <RoundtripTimes {...{ maxRoundTripTime, minRoundTripTime, medianRoundTripTime }} />
              ) : (
                <EmptyState title={sharedMessages.noData} message={m.noRoundtrip} />
              )}
            </div>
          </div>
          <hr className={style.gtwStatusPanelDivider} />
          <div className={style.gtwStatusPanelLowerContainer}>
            <div className="w-full">
              <SectionTitle title={m.transmissions} tooltip={m.transmissionsTooltip} />
              <Transmissions gatewayStats={gatewayStats} isDisconnected={isDisconnected} />
            </div>
            <div className="w-full">
              <SectionTitle
                title={m.dutyCycleUtilization}
                tooltip={m.dutyCycleUtilizationTooltip}
              />
              <div>
                {showDutyCycleUtilization ? (
                  gatewayStats.sub_bands.map((band, index) => (
                    <DutyCycleUtilization key={index} band={band} />
                  ))
                ) : (
                  <EmptyState title={sharedMessages.noData} message={m.noDutyCycle} />
                )}
              </div>
            </div>
          </div>
        </>
      )}
    </Panel>
  )
}

export default GatewayStatusPanel
