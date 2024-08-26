// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

import React, { useCallback, useRef } from 'react'
import { NavLink, useLocation } from 'react-router-dom'
import classnames from 'classnames'

import { PAGE_SIZES } from '@ttn-lw/constants/page-sizes'

import Icon, { IconHome, IconApplication, IconGateway } from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import getCookie from '@console/lib/table-utils'

import style from './switcher.styl'

const Switcher = ({ isMinimized }) => {
  const overviewRef = useRef(null)
  const applicationsRef = useRef(null)
  const gatewaysRef = useRef(null)
  const { pathname } = useLocation()

  const getNavLinkClass = useCallback(
    ({ isActive }) =>
      classnames(style.link, {
        [style.active]: isActive,
      }),
    [],
  )

  const getOverviewNavLinkClass = classnames(style.link, {
    [style.active]: !pathname.startsWith('/applications') && !pathname.startsWith('/gateways'),
  })

  const appPageSize = getCookie('applications-list-page-size')
  const appParam = `?page-size=${appPageSize ? appPageSize : PAGE_SIZES.REGULAR}`
  const gtwPageSize = getCookie('gateways-list-page-size')
  const gtwParam = `?page-size=${gtwPageSize ? gtwPageSize : PAGE_SIZES.REGULAR}`

  return (
    <div
      className={classnames(style.switcherContainer, {
        [style.isMinimized]: isMinimized,
      })}
    >
      <NavLink to="/" className={getOverviewNavLinkClass} ref={overviewRef}>
        <Icon icon={IconHome} className={style.icon} />
        <Message className={style.caption} content={sharedMessages.home} />
      </NavLink>
      <NavLink to={`/applications${appParam}`} className={getNavLinkClass} ref={applicationsRef}>
        <Icon icon={IconApplication} className={style.icon} />
        <Message className={style.caption} content={sharedMessages.applications} />
      </NavLink>
      <NavLink to={`/gateways${gtwParam}`} className={getNavLinkClass} ref={gatewaysRef}>
        <Icon icon={IconGateway} className={style.icon} />
        <Message className={style.caption} content={sharedMessages.gateways} />
      </NavLink>
    </div>
  )
}

Switcher.propTypes = {
  isMinimized: PropTypes.bool,
}

Switcher.defaultProps = {
  isMinimized: false,
}

export default Switcher
