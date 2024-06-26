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

import React, { useCallback, useEffect, useRef, useState } from 'react'
import { defineMessages, FormattedNumber } from 'react-intl'
import classnames from 'classnames'
import clipboard from 'clipboard'
import { useDispatch, useSelector } from 'react-redux'
import { createSelector } from 'reselect'

import { STACK_COMPONENTS_MAP } from '@ttn-lw/../../sdk/js/dist/util/constants'

import Icon, {
  IconCodeDots,
  IconArrowsMaximize,
  IconX,
  IconCopy,
  IconCopyCheck,
  IconAccessPoint,
  IconArrowNarrowUp,
  IconPlayerPause,
  IconPhotoOff,
} from '@ttn-lw/components/icon'
import Panel, { PanelError } from '@ttn-lw/components/panel'
import CodeEditor from '@ttn-lw/components/code-editor'
import Button from '@ttn-lw/components/button'
import PortalledModal from '@ttn-lw/components/modal/portalled'
import Link from '@ttn-lw/components/link'
import Spinner from '@ttn-lw/components/spinner'

import Message from '@ttn-lw/lib/components/message'

import LastSeen from '@console/components/last-seen'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import { getDevice } from '@console/store/actions/devices'

import {
  selectDeviceModelById,
  selectDeviceModelFetching,
} from '@console/store/selectors/device-repository'
import { selectDeviceByIds, selectDeviceFetching } from '@console/store/selectors/devices'

import style from './latest-decoded-payload-panel.styl'

const deviceNameAndImageSelector = createSelector(
  [
    (state, appId, devId) => selectDeviceByIds(state, appId, devId)?.name,
    (state, appId, devId) => {
      const device = selectDeviceByIds(state, appId, devId)
      return device
        ? selectDeviceModelById(state, device.version_ids?.brand_id, device.version_ids?.model_id)
            ?.photos?.main
        : undefined
    },
  ],
  (deviceName, image) => ({ deviceName, image }),
)

const m = defineMessages({
  latestDecodedPayload: 'Latest decoded payload',
  source: 'Source: {source}',
  received: 'Received',
  seeInLiveData: 'See in live data',
  up: '{up} up',
  rssi: `{rssi}dBm RSSI`,
  snr: `{snr}dB SNR`,
})

const hasDecodedPayload = data => {
  const type = data?.['@type']?.split('.')?.pop()

  return (
    type === 'ApplicationUplink' ||
    type === 'ApplicationUplinkNormalized' ||
    type === 'ApplicationUp'
  )
}

const LatestDecodedPayloadPanel = ({ appId, events, shortCutLinkPath }) => {
  const [selectedEvent, setSelectedEvent] = useState(null)
  const [copied, setCopied] = useState(false)
  const [isHovered, setIsHovered] = useState(false)
  const [latestEvent, setLatestEvent] = useState(null)
  const dispatch = useDispatch()

  const _timer = useRef(null)
  const containerElem = useRef(null)
  const copyElem = useRef(null)

  const actualLastEvent = events.find(e => hasDecodedPayload(e.data))
  // Save latestEvent only if it there are 10 seconds between actualLastEvent and lastEvent (throttling).
  if (
    actualLastEvent &&
    (!latestEvent || Date.parse(actualLastEvent.time) - Date.parse(latestEvent.time) > 10000) &&
    // Do not update latestEvent if it is hovered or selected.
    !isHovered &&
    !selectedEvent
  ) {
    setLatestEvent(actualLastEvent)
  }

  const formattedPayload = JSON.stringify(
    latestEvent?.data.uplink_message?.decoded_payload ?? {},
    null,
    2,
  )

  const devId = latestEvent ? latestEvent.identifiers[0].device_ids.device_id : null

  const { deviceName, image } = useSelector(state =>
    deviceNameAndImageSelector(state, appId, devId),
  )
  const imageFetching = useSelector(
    state => selectDeviceModelFetching(state) || selectDeviceFetching(state),
  )

  // Fetch device data when payload event is received.
  useEffect(() => {
    if (devId) {
      dispatch(
        attachPromise(
          getDevice(appId, devId, ['name'], {
            cache: true,
            startStream: false,
            fetchModel: true,
            // Only fetch from IS and AS, otherwise NS would also be called unnecessarily.
            components: [STACK_COMPONENTS_MAP.is, STACK_COMPONENTS_MAP.as],
            modelSelector: ['photos'],
          }),
        ),
      )
    }
  }, [appId, devId, dispatch])

  useEffect(
    () => () => {
      clearTimeout(_timer.current)
    },
    [],
  )

  const handleMouseEnter = useCallback(() => {
    setIsHovered(true)
  }, [])

  const handleMouseLeave = useCallback(() => {
    setIsHovered(false)
  }, [])

  const handleOpenMaximizeCodeModal = useCallback(() => {
    setSelectedEvent(latestEvent)
  }, [latestEvent])

  const handleCloseMaximizeCodeModal = useCallback(() => {
    setSelectedEvent(null)
  }, [])

  const handleCopyClick = useCallback(() => {
    if (copied) {
      return
    }
    setCopied(true)

    _timer.current = setTimeout(() => {
      setCopied(false)
    }, 3000)
  }, [copied])

  useEffect(() => {
    if (copyElem && copyElem.current) {
      new clipboard(copyElem.current, { container: containerElem.current })
    }
  }, [])

  const getContent = useCallback(
    (event, minLines = 3) =>
      event && (
        <>
          <Link
            to={`devices/${devId}`}
            className={classnames(style.header, 'd-flex j-between p-cs-m')}
          >
            <div className="d-inline-flex al-center gap-cs-xs c-text-neutral-heavy">
              <div className={style.imageWrapper}>
                {imageFetching ? (
                  <Spinner className={style.spinner} after={0} micro center faded />
                ) : image ? (
                  <img className={style.deviceImage} alt={deviceName} src={image} />
                ) : (
                  <Icon icon={IconPhotoOff} className={style.deviceIcon} />
                )}
              </div>
              <div className="flex-column">
                <span className="fw-bold">{deviceName || devId}</span>
                <div className="d-inline-flex al-center gap-cs-xs">
                  <div className="d-inline-flex al-center gap-cs-xxs">
                    <Icon icon={IconAccessPoint} className="c-icon-neutral-normal" />
                    <Message
                      content={m.rssi}
                      className="c-text-neutral-semilight"
                      values={{
                        rssi: event?.data.uplink_message?.rx_metadata?.[0]?.rssi ?? 0,
                      }}
                    />
                  </div>
                  <div className="d-inline-flex al-center gap-cs-xxs">
                    <Icon icon={IconAccessPoint} className="c-icon-neutral-normal" />
                    <Message
                      content={m.snr}
                      className="c-text-neutral-semilight"
                      values={{
                        snr: event?.data.uplink_message?.rx_metadata?.[0]?.snr ?? 0,
                      }}
                    />
                  </div>
                </div>
              </div>
            </div>
            <div className={style.rightHeaderColumn}>
              <LastSeen
                statusClassName={style.receivedStatus}
                message={m.received}
                lastSeen={event?.time}
                short
                displayMessage
                className="c-text-neutral-semilight"
              />
              <div className="d-inline-flex al-center gap-cs-xxs">
                <Icon icon={IconArrowNarrowUp} className="c-icon-neutral-normal" />
                <Message
                  component="span"
                  content={m.up}
                  className="c-text-neutral-semilight"
                  values={{
                    up: <FormattedNumber value={event?.data.uplink_message?.f_cnt ?? 0} />,
                  }}
                />
              </div>
            </div>
          </Link>
          <div className="pos-relative">
            <div className={style.cornerIcons} ref={containerElem}>
              <Button
                icon={copied ? IconCopyCheck : IconCopy}
                className={style.maximize}
                data-clipboard-text={formattedPayload}
                onClick={handleCopyClick}
                ref={copyElem}
                naked
                small
              />
              {!Boolean(selectedEvent) && (
                <Button
                  naked
                  icon={IconArrowsMaximize}
                  small
                  className={style.maximize}
                  onClick={handleOpenMaximizeCodeModal}
                />
              )}
            </div>
            <CodeEditor
              className={style.codeWrapper}
              value={formattedPayload}
              language="json"
              name="latest_decoded_payload"
              maxLines={Infinity}
              minLines={minLines}
              readOnly
            />
          </div>
        </>
      ),
    [
      copied,
      devId,
      deviceName,
      formattedPayload,
      handleCopyClick,
      handleOpenMaximizeCodeModal,
      image,
      imageFetching,
      selectedEvent,
    ],
  )

  return (
    <Panel
      title={m.latestDecodedPayload}
      icon={isHovered ? IconPlayerPause : IconCodeDots}
      shortCutLinkTitle={m.seeInLiveData}
      shortCutLinkPath={`${shortCutLinkPath}${latestEvent ? `?eventId=${latestEvent?.eventId}` : ''}`}
      className={style.panel}
      onMouseEnter={handleMouseEnter}
      onMouseLeave={handleMouseLeave}
    >
      {latestEvent ? (
        <div className="pos-relative">
          {getContent(latestEvent)}
          <PortalledModal
            visible={Boolean(selectedEvent)}
            noTitleLine
            noControlBar
            className={style.modalBody}
          >
            <div className="w-full">
              <div className="d-flex j-between al-center mb-cs-xl gap-cs-m">
                <div className="d-flex gap-cs-xs al-center overflow-hidden">
                  <Icon icon={IconCodeDots} className={style.headerIcon} />
                  <Message content={m.latestDecodedPayload} className={style.headerTitle} />
                </div>
                <Button naked icon={IconX} onClick={handleCloseMaximizeCodeModal} />
              </div>

              {getContent(selectedEvent)}
              <div className="d-flex j-center al-center gap-cs-m pt-cs-xl">
                <Button
                  secondary
                  icon={IconX}
                  message={sharedMessages.close}
                  onClick={handleCloseMaximizeCodeModal}
                />
              </div>
            </div>
          </PortalledModal>
        </div>
      ) : (
        <PanelError>
          <Message component="p" content={sharedMessages.noRecentActivity} />
        </PanelError>
      )}
    </Panel>
  )
}

LatestDecodedPayloadPanel.propTypes = {
  appId: PropTypes.string.isRequired,
  events: PropTypes.events.isRequired,
  shortCutLinkPath: PropTypes.string.isRequired,
}

export default LatestDecodedPayloadPanel
