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

import { IconCodeDots } from '@ttn-lw/components/icon'
import Panel, { PanelError } from '@ttn-lw/components/panel'
import CodeEditor from '@ttn-lw/components/code-editor'

import Message from '@ttn-lw/lib/components/message'

import LastSeen from '@console/components/last-seen'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import style from './latest-decoded-payload-panel.styl'

const m = defineMessages({
  latestDecodedPayload: 'Latest decoded payload',
  source: 'Source: {source}',
  received: 'Received',
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

const LatestDecodedPayloadPanel = ({ events }) => {
  const [latestEvent, setLatestEvent] = useState(null)

  const throttledSetLatestEvent = useCallback(
    throttle(newEvent => setLatestEvent(newEvent), 10000),
    [],
  )

  useEffect(() => {
    const normalizedEvents = events
      .map(e => ({ time: e.time, decodedPayload: findKey(e.data, 'decoded_payload') }))
      .filter(e => !!e.decodedPayload)
    if (normalizedEvents.length > 0) {
      throttledSetLatestEvent(normalizedEvents[0])
    }
  }, [events, throttledSetLatestEvent])

  return (
    <Panel
      title={m.latestDecodedPayload}
      icon={IconCodeDots}
      shortCutLinkTitle={sharedMessages.expand}
      className={style.panel}
      shortCutLinkDisabled={!latestEvent}
    >
      {latestEvent ? (
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
              lastSeen={latestEvent.time}
              className={style.received}
            />
          </div>
          <CodeEditor
            value={JSON.stringify(latestEvent.decodedPayload, null, 2)}
            language="json"
            name="latest_decoded_payload"
            maxLines={12}
            readOnly
          />
        </>
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
}

export default LatestDecodedPayloadPanel
