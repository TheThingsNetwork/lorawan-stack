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

import React, { useCallback, useEffect, useState } from 'react'
import { defineMessages } from 'react-intl'
import { forOwn, isObject, throttle } from 'lodash'

import Icon, { IconCodeDots, IconArrowsMaximize, IconX, IconCopy } from '@ttn-lw/components/icon'
import Panel, { PanelError } from '@ttn-lw/components/panel'
import CodeEditor from '@ttn-lw/components/code-editor'
import Button from '@ttn-lw/components/button'
import PortalledModal from '@ttn-lw/components/modal/portalled'
import toast from '@ttn-lw/components/toast'

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
})

const findKey = (obj, keyToFind) => {
  let result

  const recursiveSearch = obj => {
    forOwn(obj, (value, key) => {
      if (key === keyToFind) {
        result = value
        return false // Exit early
      } else if (isObject(value)) {
        result = recursiveSearch(value)
        if (result) return false // Exit early if found
      }
    })
    return result
  }

  return recursiveSearch(obj)
}

const LatestDecodedPayloadPanel = ({ events, shortCutLinkPath }) => {
  const [latestEvent, setLatestEvent] = useState(null)
  const [selectedEvent, setSelectedEvent] = useState(null)

  const throttledSetLatestEvent = useCallback(
    throttle(newEvent => setLatestEvent(newEvent), 10000),
    [],
  )

  useEffect(() => {
    const normalizedEvents = events
      .map(e => ({
        eventId: e.unique_id,
        time: e.time,
        decodedPayload: findKey(e.data, 'decoded_payload'),
      }))
      .filter(e => !!e.decodedPayload)
    if (normalizedEvents.length > 0) {
      throttledSetLatestEvent(normalizedEvents[0])
    }
  }, [events, throttledSetLatestEvent])

  const handleOpenMaximizeCodeModal = useCallback(() => {
    setSelectedEvent(latestEvent)
  }, [latestEvent])

  const handleCloseMaximizeCodeModal = useCallback(() => {
    setSelectedEvent(null)
  }, [])

  const handleCopy = useCallback(() => {
    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(JSON.stringify(selectedEvent?.decodedPayload)).then(() => {
        toast({
          title: m.success,
          message: sharedMessages.copiedToClipboard,
          type: toast.types.SUCCESS,
        })
      })
    }
  }, [selectedEvent])

  const getContent = useCallback(
    (event, maxLines = 8, minLines = undefined) => (
      <>
        <div className="d-flex j-between pt-cs-m pb-cs-m">
          <Message
            uppercase
            content={m.source}
            values={{ source: sharedMessages.liveData.defaultMessage }}
            className={style.source}
          />
          <LastSeen
            statusClassName={style.receivedStatus}
            message={m.received}
            lastSeen={event?.time}
            className={style.received}
          />
        </div>

        <CodeEditor
          value={JSON.stringify(event?.decodedPayload, null, 2)}
          language="json"
          name="latest_decoded_payload"
          maxLines={maxLines}
          minLines={minLines}
          readOnly
        />
      </>
    ),
    [],
  )

  return (
    <Panel
      title={m.latestDecodedPayload}
      icon={IconCodeDots}
      shortCutLinkTitle={m.seeInLiveData}
      shortCutLinkPath={`${shortCutLinkPath}?eventId=${latestEvent?.eventId}`}
      className={style.panel}
      shortCutLinkDisabled={!latestEvent}
    >
      {latestEvent ? (
        <div className="pos-relative">
          {getContent(latestEvent)}
          <Button
            naked
            icon={IconArrowsMaximize}
            small
            className={style.maximize}
            onClick={handleOpenMaximizeCodeModal}
          />
          <PortalledModal visible={selectedEvent} noTitleLine noControlLine>
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
                  icon={IconCopy}
                  message={sharedMessages.copy}
                  onClick={handleCopy}
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
  events: PropTypes.events.isRequired,
  shortCutLinkPath: PropTypes.string.isRequired,
}

export default LatestDecodedPayloadPanel
