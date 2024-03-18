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

import React, { forwardRef } from 'react'
import classnames from 'classnames'
import PropTypes from 'prop-types'
import * as Icons from '@tabler/icons-react'

import TtsIcon from './replacements/tts-icon'

import style from './icon.styl'

const hyphenCaseToPascalCase = str =>
  str
    .split('-')
    .map(word => word.charAt(0).toUpperCase() + word.slice(1))
    .join('')

// A map of hardcoded names to their corresponding icons.
// Keep these sorted alphabetically.
const hardcoded = {
  access: 'lock',
  'admin-panel': 'user-shield',
  'api-keys': 'key',
  application: 'device-desktop-analytics',
  cluster: 'world',
  collaborators: 'users',
  develop: 'code',
  device: 'cpu',
  downlink: 'arrow-down',
  event: 'info',
  'event-clear-all': 'clear-all',
  'event-connection': 'transfer',
  'event-create': 'circle-plus',
  'event-delete': 'trash',
  'event-downlink': 'arrow-down',
  'event-error': 'exclamation-circle',
  'event-gateway-connect': 'bolt',
  'event-gateway-disconnect': 'bolt-off',
  'event-join': 'circles-relation',
  'event-mode': 'adjustments-horizontal',
  'event-rekey': 'key',
  'event-status': 'heartbeat',
  'event-switch': 'switch',
  'event-update': 'edit',
  'event-uplink': 'arrow-up',
  'expand-down': 'arrow-down',
  'expand-up': 'arrow-up',
  gateway: 'router',
  'general-settings': 'settings',
  'import-devices': 'playlist-add',
  integration: 'arrow-merge-alt-right',
  join: 'circles-relation',
  'live-data': 'article',
  location: 'map-pin',
  organization: 'users-group',
  overview: 'laoyout-dashboard',
  'packet-broker': 'aperture',
  'payload-format': 'source-code',
  'oauth-clients': 'brand-oauth',
  support: 'lifebuoy',
  sort: 'selector',
  'sort-order-asc': 'sort-ascending',
  'sort-order-desc': 'sort-descending',
  uplink: 'arrow-up',
  'user-management': 'user-cog',
  valid: 'circle-check',
}

const replaced = {
  tts: TtsIcon,
}

const Icon = forwardRef((props, ref) => {
  const {
    icon,
    className,
    nudgeUp,
    nudgeDown,
    small,
    large,
    textPaddedLeft,
    textPaddedRight,
    size,
    ...rest
  } = props

  const classname = classnames(className, {
    [style.nudgeUp]: nudgeUp,
    [style.nudgeDown]: nudgeDown,
    [style.large]: large,
    [style.small]: small,
    [style.textPaddedLeft]: textPaddedLeft,
    [style.textPaddedRight]: textPaddedRight,
  })

  const Icon = Icons[`Icon${hyphenCaseToPascalCase(hardcoded[icon] || icon)}`] || replaced[icon]

  if (!Icon) {
    console.warn(
      `Icon${hyphenCaseToPascalCase(hardcoded[icon] || icon)} (${icon}) is not available in the tabler icon set`,
    )
    return null
  }

  return <Icon className={classname} ref={ref} {...rest} size={small ? 16 : size} />
})

Icon.propTypes = {
  className: PropTypes.string,
  /** Which icon to display, using tabler icon set. */
  icon: PropTypes.string.isRequired,
  /** Renders a bigger icon. */
  large: PropTypes.bool,
  /** Nudges the icon down by one pixel using position: relative. */
  nudgeDown: PropTypes.bool,
  /** Nudges the icon up by one pixel using position: relative. */
  nudgeUp: PropTypes.bool,
  /** The size of the icon. */
  size: PropTypes.number,
  /** Renders a smaller icon. */
  small: PropTypes.bool,
  /** Whether icon should be padded for a text displayed left to it. */
  textPaddedLeft: PropTypes.bool,
  /** Whether icon should be padded for a text displayed right to it. */
  textPaddedRight: PropTypes.bool,
}

Icon.defaultProps = {
  className: undefined,
  large: false,
  nudgeDown: false,
  nudgeUp: false,
  size: 20,
  small: false,
  textPaddedLeft: false,
  textPaddedRight: false,
}

export default Icon
