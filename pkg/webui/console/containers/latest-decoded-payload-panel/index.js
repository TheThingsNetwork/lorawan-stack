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
import { throttle } from 'lodash'
import classnames from 'classnames'

import tts from '@console/api/tts'
import deviceIcon from '@assets/misc/end-device.svg'

import Icon, {
  IconCodeDots,
  IconArrowsMaximize,
  IconX,
  IconCopy,
  IconCopyCheck,
  IconAccessPoint,
  IconArrowNarrowUp,
} from '@ttn-lw/components/icon'
import Panel, { PanelError } from '@ttn-lw/components/panel'
import CodeEditor from '@ttn-lw/components/code-editor'
import Button from '@ttn-lw/components/button'
import PortalledModal from '@ttn-lw/components/modal/portalled'

import Message from '@ttn-lw/lib/components/message'

import LastSeen from '@console/components/last-seen'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import style from './latest-decoded-payload-panel.styl'

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
  const [latestEvent, setLatestEvent] = useState(null)
  const [selectedEvent, setSelectedEvent] = useState(null)
  const [copied, setCopied] = useState(false)

  const _timer = useRef(null)

  const throttledSetLatestEvent = useCallback(
    throttle(async (newEvent, devId) => {
      const { name, version_ids } = await tts.Applications.Devices.getById(appId, devId, [
        'name',
        'version_ids',
      ])
      const hasModel = Boolean(version_ids?.brand_id) && Boolean(version_ids?.model_id)
      let picture = deviceIcon
      if (hasModel) {
        const { photos } = await tts.Applications.Devices.Repository.getModel(
          appId,
          version_ids.brand_id,
          version_ids.model_id,
          ['photos'],
        )
        picture = photos.main
      }

      return setLatestEvent({ ...newEvent, deviceName: name, picture })
    }, 10000),
    [],
  )

  useEffect(
    () => () => {
      clearTimeout(_timer.current)
    },
    [],
  )

  useEffect(() => {
    if (!events.length) return

    const firstEventWithPayload = events.find(e => hasDecodedPayload(e.data))

    if (firstEventWithPayload) {
      throttledSetLatestEvent(
        {
          eventId: firstEventWithPayload.unique_id,
          time: firstEventWithPayload.time,
          decodedPayload: firstEventWithPayload.data.uplink_message?.decoded_payload ?? {},
          rssi: firstEventWithPayload.data.uplink_message?.rx_metadata?.[0]?.rssi ?? 0,
          snr: firstEventWithPayload.data.uplink_message?.rx_metadata?.[0]?.snr ?? 0,
          numUplink: firstEventWithPayload.data.uplink_message?.f_cnt ?? 0,
        },
        firstEventWithPayload.data.end_device_ids.device_id,
      )
    }
  }, [events, throttledSetLatestEvent])

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
    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(JSON.stringify(selectedEvent?.decodedPayload, null, 2))
    }
    setCopied(true)

    _timer.current = setTimeout(() => {
      setCopied(false)
    }, 3000)
  }, [copied, selectedEvent])

  const getContent = useCallback(
    (event, maxLines = 8, minLines = 3) => (
      <>
        <div className={classnames(style.header, 'd-flex j-between p-cs-m')}>
          <div className="d-inline-flex al-center gap-cs-xs">
            <img className={style.deviceImage} alt={event?.deviceName} src={event?.picture} />
            <div className="flex-column">
              <Message content={event?.deviceName} className="fw-bold" />
              <div className="d-inline-flex al-center gap-cs-xs">
                <div className="d-inline-flex al-center gap-cs-xxs">
                  <Icon icon={IconAccessPoint} />
                  <Message
                    content={m.rssi}
                    className="c-text-neutral-semilight"
                    values={{
                      rssi: event?.rssi,
                    }}
                  />
                </div>
                <div className="d-inline-flex al-center gap-cs-xxs">
                  <Icon icon={IconAccessPoint} />
                  <Message
                    content={m.snr}
                    className="c-text-neutral-semilight"
                    values={{
                      snr: event?.snr,
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
              <Icon icon={IconArrowNarrowUp} />
              <Message
                component="span"
                content={m.up}
                className="c-text-neutral-semilight"
                values={{
                  up: <FormattedNumber value={event?.numUplink} />,
                }}
              />
            </div>
          </div>
        </div>
        {!selectedEvent && (
          <Button
            naked
            icon={IconArrowsMaximize}
            small
            className={style.maximize}
            onClick={handleOpenMaximizeCodeModal}
          />
        )}
        <CodeEditor
          className={style.codeWrapper}
          value={JSON.stringify(event?.decodedPayload, null, 2)}
          language="json"
          name="latest_decoded_payload"
          maxLines={maxLines}
          minLines={minLines}
          readOnly
        />
      </>
    ),
    [handleOpenMaximizeCodeModal, selectedEvent],
  )

  return (
    <Panel
      title={m.latestDecodedPayload}
      icon={IconCodeDots}
      shortCutLinkTitle={m.seeInLiveData}
      shortCutLinkPath={`${shortCutLinkPath}${latestEvent ? `?eventId=${latestEvent?.eventId}` : ''}`}
      className={style.panel}
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

              {getContent(selectedEvent, 20)}
              <div className="d-flex j-center al-center gap-cs-m pt-cs-xl">
                <Button
                  secondary
                  icon={copied ? IconCopyCheck : IconCopy}
                  message={copied ? sharedMessages.copiedToClipboard : sharedMessages.copy}
                  onClick={handleCopyClick}
                />
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
