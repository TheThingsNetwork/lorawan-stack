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
import classnames from 'classnames'
import PropTypes from 'prop-types'

import style from './icon.styl'

// A map of hardcoded names to their corresponding icons.
const hardcoded = {
  devices: 'device_hub',
  device: 'device_hub',
  settings: 'tune',
  integration: 'settings_ethernet',
  data: 'poll',
  sort: 'arrow_drop_down',
  overview: 'apps',
  application: 'web_asset',
  gateway: 'router',
  organization: 'people',
  api_keys: 'lock',
  link: 'code',
  payload_formats: 'code',
  develop: 'code',
  access: 'lock',
  general_settings: 'settings',
  location: 'place',
  user: 'person',
  event: 'wifi',
  event_create: 'add_circle',
  event_delete: 'delete',
  event_update: 'edit',
  event_uplink: 'arrow_drop_up',
  event_downlink: 'arrow_drop_down',
  uplink: 'trending_up',
  downlink: 'trending_down',
}

const Icon = function ({
  icon = '',
  className,
  nudgeUp = false,
  nudgeDown = false,
  small = false,
  large = false,
  ...rest
}) {

  const classname = classnames(style.icon, className, {
    [style.nudgeUp]: nudgeUp,
    [style.nudgeDown]: nudgeDown,
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

Icon.propTypes = {
  /** Which icon to display, using google material icon set */
  icon: PropTypes.string.isRequired,
  /** Nudges the icon up by one pixel using position: relative */
  nudgeUp: PropTypes.bool,
  /** Nudges the icon down by one pixel using position: relative */
  nudgeDown: PropTypes.bool,
  /** Renders a smaller icon */
  small: PropTypes.bool,
  /** Renders a bigger icon */
  large: PropTypes.bool,
}

export default Icon
