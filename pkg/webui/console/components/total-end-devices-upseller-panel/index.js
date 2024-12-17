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

import React from 'react'

import { IconBolt, IconDevice } from '@ttn-lw/components/icon'
import Panel from '@ttn-lw/components/panel'
import Button from '@ttn-lw/components/button'

import Message from '@ttn-lw/lib/components/message'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import style from './total-end-devices-upseller-panel.styl'

const TotalEndDevicesUpsellerPanel = () => {
  const upgradeUrl = 'https://www.thethingsindustries.com/stack/plans/'

  return (
    <Panel
      title={sharedMessages.totalEndDevices}
      icon={IconDevice}
      shortCutLinkButton
      shortCutLinkPath="#"
      shortCutLinkDisabled
      className={style.panel}
      compact
    >
      <div className={style.upseller}>
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
    </Panel>
  )
}

export default TotalEndDevicesUpsellerPanel
