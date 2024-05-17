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

import React, { useCallback, useState } from 'react'
import classnames from 'classnames'

import { IconChartBar, IconBolt } from '@ttn-lw/components/icon'
import Panel from '@ttn-lw/components/panel'
import Toggle from '@ttn-lw/components/panel/toggle'
import Button from '@ttn-lw/components/button'

import Message from '@ttn-lw/lib/components/message'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import style from './blurry-network-activity-panel.styl'

const toggleOptions = [
  { label: sharedMessages.packetsPerDataRate, value: 0 },
  { label: sharedMessages.packetsPerChannel, value: 1 },
]

const BlurryNetworkActivityPanel = () => {
  const [activeToggle, setActiveToggle] = useState(0)

  const handleToggleChange = useCallback((_, value) => {
    setActiveToggle(value)
  }, [])

  return (
    <Panel
      title={sharedMessages.networkActivity}
      icon={IconChartBar}
      shortCutLinkTitle={sharedMessages.noc}
      shortCutLinkPath="#"
      shortCutLinkTarget="_blank"
      shortCutLinkDisabled
      className={style.panel}
    >
      <Toggle
        options={toggleOptions}
        active={activeToggle}
        onToggleChange={handleToggleChange}
        fullWidth
      />
      <div className={style.content}>
        <div
          className={classnames(style.upseller, {
            [style.small]: activeToggle === 0,
          })}
        >
          <Message
            className="c-text-neutral-heavy fw-bold fs-l text-center"
            content={sharedMessages.unlockTheNoc}
          />
          <Message
            className="c-text-neutral-light fs-m text-center mb-cs-l"
            content={sharedMessages.quicklyTroubleshoot}
          />
          <Button.AnchorLink
            primary
            message={sharedMessages.upgradeNow}
            icon={IconBolt}
            href="https://www.thethingsindustries.com/stack/plans/"
            target="_blank"
          />
        </div>
      </div>
    </Panel>
  )
}

export default BlurryNetworkActivityPanel
