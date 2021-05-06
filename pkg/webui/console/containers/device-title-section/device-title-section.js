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

import React from 'react'
import { defineMessages } from 'react-intl'

import deviceIcon from '@assets/misc/end-device.svg'

import Status from '@ttn-lw/components/status'
import Spinner from '@ttn-lw/components/spinner'

import Message from '@ttn-lw/lib/components/message'

import EntityTitleSection from '@console/components/entity-title-section'
import LastSeen from '@console/components/last-seen'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import style from './device-title-section.styl'

const m = defineMessages({
  lastSeenUnavailable: 'Last seen info unavailable',
})

const { Content } = EntityTitleSection

const DeviceTitleSection = props => {
  const {
    devId,
    fetching,
    device,
    uplinkFrameCount,
    downlinkFrameCount,
    lastSeen,
    children,
  } = props

  const showLastSeen = Boolean(lastSeen)
  const showUplinkCount = typeof uplinkFrameCount === 'number'
  const showDownlinkCount = typeof downlinkFrameCount === 'number'
  const notAvailableElem = <Message content={sharedMessages.notAvailable} />

  return (
    <EntityTitleSection
      id={devId}
      name={device.name}
      icon={deviceIcon}
      iconAlt={sharedMessages.device}
    >
      <Content className={style.content} creationDate={device.created_at}>
        {fetching ? (
          <Spinner after={0} faded micro inline>
            <Message content={sharedMessages.fetching} />
          </Spinner>
        ) : (
          <>
            {showLastSeen ? (
              <Status status="good" flipped>
                <LastSeen lastSeen={lastSeen} />
              </Status>
            ) : (
              <Status status="mediocre" label={m.lastSeenUnavailable} flipped />
            )}
            <Content.MessagesCount
              icon="uplink"
              value={showUplinkCount ? uplinkFrameCount : notAvailableElem}
              tooltipMessage={sharedMessages.uplinkFrameCount}
              iconClassName={showUplinkCount ? style.messageIcon : style.notAvailable}
            />
            <Content.MessagesCount
              icon="downlink"
              value={showDownlinkCount ? downlinkFrameCount : notAvailableElem}
              tooltipMessage={sharedMessages.downlinkFrameCount}
              iconClassName={showUplinkCount ? style.messageIcon : style.notAvailable}
            />
          </>
        )}
      </Content>
      {children}
    </EntityTitleSection>
  )
}

DeviceTitleSection.propTypes = {
  children: PropTypes.oneOfType([PropTypes.arrayOf(PropTypes.node), PropTypes.node]),
  devId: PropTypes.string.isRequired,
  device: PropTypes.device.isRequired,
  downlinkFrameCount: PropTypes.number,
  fetching: PropTypes.bool,
  lastSeen: PropTypes.string,
  uplinkFrameCount: PropTypes.number,
}

DeviceTitleSection.defaultProps = {
  uplinkFrameCount: undefined,
  lastSeen: undefined,
  children: null,
  fetching: false,
  downlinkFrameCount: undefined,
}

export default DeviceTitleSection
