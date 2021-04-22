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

import React, { forwardRef } from 'react'
import classnames from 'classnames'
import PropTypes from 'prop-types'

import style from './icon.styl'

// A map of hardcoded names to their corresponding icons.
// Keep these sorted alphabetically.
const hardcoded = {
  access: 'lock',
  api_keys: 'vpn_key',
  application: 'web_asset',
  collaborators: 'people',
  data: 'poll',
  develop: 'code',
  device: 'device_hub',
  devices: 'device_hub',
  downlink: 'arrow_downward',
  event: 'info',
  event_clear_all: 'clear_all',
  event_connection: 'settings_ethernet',
  event_create: 'add_circle',
  event_delete: 'delete',
  event_downlink: 'arrow_downward',
  event_error: 'error',
  event_gateway_connect: 'flash_on',
  event_gateway_disconnect: 'flash_off',
  event_join: 'link',
  event_mode: 'tune',
  event_rekey: 'vpn_key',
  event_status: 'network_check',
  event_switch: 'tune',
  event_update: 'edit',
  event_uplink: 'arrow_upward',
  expand_down: 'keyboard_arrow_down',
  expand_up: 'keyboard_arrow_up',
  gateway: 'router',
  general_settings: 'settings',
  import_devices: 'playlist_add',
  integration: 'call_merge',
  join: 'link',
  link: 'link',
  location: 'place',
  logout: 'power_settings_new',
  organization: 'people',
  overview: 'dashboard',
  payload_formats: 'code',
  settings: 'tune',
  sort_order_asc: 'arrow_drop_down',
  sort_order_desc: 'arrow_drop_up',
  uplink: 'arrow_upward',
  user_management: 'how_to_reg',
  user: 'person',
  valid: 'check_circle',
}

const Icon = forwardRef((props, ref) => {
  const { icon, className, nudgeUp, nudgeDown, small, large, ...rest } = props

  const classname = classnames(style.icon, className, {
    [style.nudgeUp]: nudgeUp,
    [style.nudgeDown]: nudgeDown,
    [style.large]: large,
    [style.small]: small,
  })

  return (
    <span className={classname} ref={ref} {...rest}>
      {hardcoded[icon] || icon}
    </span>
  )
})

Icon.propTypes = {
  className: PropTypes.string,
  /** Which icon to display, using google material icon set. */
  icon: PropTypes.string.isRequired,
  /** Renders a bigger icon. */
  large: PropTypes.bool,
  /** Nudges the icon down by one pixel using position: relative. */
  nudgeDown: PropTypes.bool,
  /** Nudges the icon up by one pixel using position: relative. */
  nudgeUp: PropTypes.bool,
  /** Renders a smaller icon. */
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
