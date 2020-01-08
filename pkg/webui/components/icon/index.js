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
  integration: 'call_merge',
  data: 'poll',
  sort_order_asc: 'arrow_drop_down',
  sort_order_desc: 'arrow_drop_up',
  overview: 'dashboard',
  application: 'web_asset',
  gateway: 'router',
  organization: 'people',
  api_keys: 'vpn_key',
  link: 'link',
  payload_formats: 'code',
  develop: 'code',
  access: 'lock',
  general_settings: 'settings',
  location: 'place',
  user: 'person',
  user_management: 'people',
  event: 'info',
  event_create: 'add_circle',
  event_delete: 'remove_circle',
  event_update: 'edit',
  uplink: 'arrow_drop_up',
  downlink: 'arrow_drop_down',
  import_devices: 'playlist_add',
  collaborators: 'people',
}

const Icon = function({ icon, className, nudgeUp, nudgeDown, small, large, ...rest }) {
  const classname = classnames(style.icon, className, {
    [style.nudgeUp]: nudgeUp,
    [style.nudgeDown]: nudgeDown,
    [style.large]: large,
    [style.small]: small,
  })

  return (
    <span className={classname} {...rest}>
      {hardcoded[icon] || icon}
    </span>
  )
}

Icon.propTypes = {
  className: PropTypes.string,
  /** Which icon to display, using google material icon set */
  icon: PropTypes.string.isRequired,
  /** Renders a bigger icon */
  large: PropTypes.bool,
  /** Nudges the icon down by one pixel using position: relative */
  nudgeDown: PropTypes.bool,
  /** Nudges the icon up by one pixel using position: relative */
  nudgeUp: PropTypes.bool,
  /** Renders a smaller icon */
  small: PropTypes.bool,
}

Icon.defaultProps = {
  className: undefined,
  large: false,
  nudgeDown: false,
  nudgeUp: false,
  small: false,
}

export default Icon
