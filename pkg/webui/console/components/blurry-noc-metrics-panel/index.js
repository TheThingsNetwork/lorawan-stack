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

import React, { useState } from 'react'
import classnames from 'classnames'

import Panel from '@ttn-lw/components/panel'
import Toggle from '@ttn-lw/components/panel/toggle'
import Button from '@ttn-lw/components/button'
import { IconBolt } from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import style from './blurry-noc-metrics-panel.styl'

const toggleOptions = [
  { label: '7 days', value: 7 },
  { label: '30 days', value: 30 },
]

const BlurryNocMetricsPanel = ({ title, icon }) => {
  const [activeToggle, setActiveToggle] = useState(7)
  const upgradeUrl = 'https://www.thethingsindustries.com/stack/plans/'

  const handleToggleChange = React.useCallback((_, value) => {
    setActiveToggle(value)
  }, [])

  return (
    <Panel
      title={title}
      icon={icon}
      shortCutLinkButton
      shortCutLinkPath="#"
      className={style.panel}
      compact
      shortCutLinkDisabled
    >
      <div className={style.content}>
        <div
          className={classnames(style.upseller, {
            [style.small]: activeToggle === 7,
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
            tertiary
            message={sharedMessages.upgradeNow}
            icon={IconBolt}
            href={upgradeUrl}
            target="_blank"
          />
        </div>
      </div>
      <Toggle
        options={toggleOptions}
        active={activeToggle}
        onToggleChange={handleToggleChange}
        fullWidth
        className="mt-cs-m"
      />
    </Panel>
  )
}

BlurryNocMetricsPanel.propTypes = {
  icon: PropTypes.icon.isRequired,
  title: PropTypes.message.isRequired,
}

export default BlurryNocMetricsPanel
