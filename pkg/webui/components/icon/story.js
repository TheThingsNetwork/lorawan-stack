// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

import Icon, {
  IconApplication,
  IconLiveData,
  IconDevice,
  IconGateway,
  IconGeneralSettings,
  IconIntegration,
  IconLock,
  IconLockOpen,
  IconMenu2,
  IconOrganization,
  IconOverview,
  IconSort,
  IconTransform,
  IconX,
} from '@ttn-lw/components/icon'

import style from './story.styl'

const icons = {
  IconDevice,
  IconIntegration,
  IconGeneralSettings,
  IconLock,
  IconLockOpen,
  IconX,
  IconMenu2,
  IconTransform,
  IconLiveData,
  IconSort,
  IconOverview,
  IconApplication,
  IconGateway,
  IconOrganization,
}

const iconElement = Object.entries(icons).map(([name, icon]) => (
  <div className={style.wrapper} key={name}>
    <Icon icon={icon} />
    {name}
  </div>
))

export default {
  title: 'Icon',
  component: Icon,
}

export const Icons = () => <div>{iconElement}</div>

export const Usage = () => (
  <div className={style.wrapper}>
    <div className={style.block}>
      <Icon icon={IconDevice} />
      <span>{'display: inline-block'}</span>
    </div>
    <br />
    <div className={style.flex}>
      <Icon icon={IconDevice} />
      <span>{'display: flex'}</span>
    </div>
  </div>
)
