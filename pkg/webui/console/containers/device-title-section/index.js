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
import { FormattedNumber, defineMessages } from 'react-intl'
import { useSelector } from 'react-redux'

import deviceIcon from '@assets/misc/end-device.svg'

import Icon, { IconHelp, IconDownlink, IconUplink } from '@ttn-lw/components/icon'
import Status from '@ttn-lw/components/status'
import Tooltip from '@ttn-lw/components/tooltip'
import DocTooltip from '@ttn-lw/components/tooltip/doc'

import Message from '@ttn-lw/lib/components/message'
import DateTime from '@ttn-lw/lib/components/date-time'

import EntityTitleSection from '@console/components/entity-title-section'
import LastSeen from '@console/components/last-seen'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import {
  selectDeviceDerivedAppDownlinkFrameCount,
  selectDeviceDerivedNwkDownlinkFrameCount,
  selectDeviceDerivedUplinkFrameCount,
  selectDeviceLastSeen,
  selectSelectedCombinedDeviceId,
  selectSelectedDevice,
} from '@console/store/selectors/devices'

const m = defineMessages({
  uplinkDownlinkTooltip:
    'The number of sent uplinks and received downlinks of this end device since the last frame counter reset.{lineBreak}App: frame counter for application downlinks (`FPort >=1`).{lineBreak}Nwk: frame counter for network downlinks (`FPort = 0`)',
  lastSeenAvailableTooltip:
    'The elapsed time since the network registered the last activity of this end device. This is determined from sent uplinks, confirmed downlinks or (re)join requests.{lineBreak}The last activity was received at {lastActivityInfo}',
  noActivityTooltip:
    'The network has not registered any activity from this end device yet. This could mean that your end device has not sent any messages yet or only messages that cannot be handled by the network, e.g. due to a mismatch of EUIs or frequencies.',
})

const { Content } = EntityTitleSection

const DeviceTitleSection = () => {
  const device = useSelector(selectSelectedDevice)
  const [appId, devId] = useSelector(selectSelectedCombinedDeviceId).split('/')
  const uplinkFrameCount = useSelector(state =>
    selectDeviceDerivedUplinkFrameCount(state, appId, devId),
  )
  const downlinkAppFrameCount = useSelector(state =>
    selectDeviceDerivedAppDownlinkFrameCount(state, appId, devId),
  )
  const downlinkNwkFrameCount = useSelector(state =>
    selectDeviceDerivedNwkDownlinkFrameCount(state, appId, devId),
  )
  const lastSeen = useSelector(state => selectDeviceLastSeen(state, appId, devId))
  const showLastSeen = Boolean(lastSeen)
  const showUplinkCount = typeof uplinkFrameCount === 'number'
  const showAppDownlinkCount = typeof downlinkAppFrameCount === 'number'
  const showNwkDownlinkCount = typeof downlinkNwkFrameCount === 'number'

  const notAvailableElem = <Message content={sharedMessages.notAvailable} />
  const downlinkValue =
    showAppDownlinkCount && showNwkDownlinkCount ? (
      <>
        <FormattedNumber value={downlinkAppFrameCount} /> {'(App) / '}
        <FormattedNumber value={downlinkNwkFrameCount} /> {'(Nwk)'}
      </>
    ) : showAppDownlinkCount ? (
      <>
        <FormattedNumber value={downlinkAppFrameCount} /> {'(App)'}
      </>
    ) : showNwkDownlinkCount ? (
      <>
        <FormattedNumber value={downlinkNwkFrameCount} /> {'(Nwk)'}
      </>
    ) : (
      notAvailableElem
    )
  const lastActivityInfo = lastSeen ? <DateTime value={lastSeen} noTitle /> : lastSeen
  const lineBreak = <br />
  const bottomBarLeft = (
    <>
      <Tooltip
        content={
          <Message
            content={m.uplinkDownlinkTooltip}
            values={{ lineBreak: <br /> }}
            convertBackticks
          />
        }
      >
        <div className="d-flex">
          <Content.MessagesCount
            icon={IconUplink}
            value={showUplinkCount ? uplinkFrameCount : notAvailableElem}
            iconClassName={
              showUplinkCount ? 'd-flex c-text-brand-normal' : 'd-flex c-text-neutral-light'
            }
          />
          <Content.MessagesCount
            icon={IconDownlink}
            value={downlinkValue}
            iconClassName={
              showUplinkCount ? 'd-flex c-text-brand-normal' : 'd-flex c-text-neutral-light'
            }
            helpTooltip
          />
        </div>
      </Tooltip>
      {showLastSeen ? (
        <DocTooltip
          docPath="/reference/last-activity"
          content={
            <Message
              content={m.lastSeenAvailableTooltip}
              values={{ lineBreak, lastActivityInfo }}
            />
          }
        >
          <LastSeen lastSeen={lastSeen} flipped noTitle>
            <Icon icon={IconHelp} textPaddedLeft small nudgeUp className="c-text-neutral-light" />
          </LastSeen>
        </DocTooltip>
      ) : (
        <DocTooltip
          docPath="/devices/troubleshooting/#my-device-wont-join-what-do-i-do"
          docTitle={sharedMessages.troubleshooting}
          content={<Message content={m.noActivityTooltip} />}
        >
          <Status status="mediocre" label={sharedMessages.noActivityYet} flipped>
            <Icon icon={IconHelp} textPaddedLeft small nudgeUp className="c-text-neutral-light" />
          </Status>
        </DocTooltip>
      )}
    </>
  )

  return (
    <EntityTitleSection
      id={devId}
      name={device.name}
      icon={deviceIcon}
      iconAlt={sharedMessages.device}
    >
      <Content
        className="m-vert-ls-xxs m-sides-0"
        creationDate={device.created_at}
        bottomBarLeft={bottomBarLeft}
      />
    </EntityTitleSection>
  )
}

export default DeviceTitleSection
