// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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
import classnames from 'classnames'

import style from './style.styl'

// A map of hardcoded names to their corresponding icons.
const hardcoded = {
  gateway: 'wifi_tethering',
  application: 'layers',
  collaborator: 'supervisor_account',
  devices: 'devices',
  settings: 'tune',
  integration: 'settings_ethernet',
  data: 'poll',
}

export default function ({
  icon = '',
  className,
  nudgeTop,
  nudgeBottom,
  small,
  large,
  ...rest
}) {

  const classname = classnames(style.icon, className, {
    [style.nudgeTop]: nudgeTop,
    [style.nudgeBottom]: nudgeBottom,
    [style.large]: large,
    [style.small]: small,
  })

  return (
    <span
      className={classname}
      {...rest}
    >
      {hardcoded[icon] || icon}
    </span>
  )
}

