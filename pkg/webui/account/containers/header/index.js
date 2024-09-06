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

import React, { useCallback } from 'react'
import { useDispatch, useSelector } from 'react-redux'

import {
  IconLogout,
  IconUserCircle,
  IconBook,
  IconAdminPanel,
  IconSupport,
} from '@ttn-lw/components/icon'
import HeaderComponent from '@ttn-lw/components/header'
import Dropdown from '@ttn-lw/components/dropdown'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import { selectAssetsRootPath, selectBrandingRootPath } from '@ttn-lw/lib/selectors/env'
import PropTypes from '@ttn-lw/lib/prop-types'

import { logout } from '@console/store/actions/logout'

import { selectUser, selectUserIsAdmin } from '@console/store/selectors/user'

import Logo from '../logo'

const Header = ({ onMenuClick, alwaysShowLogo }) => {
  const dispatch = useDispatch()

  const handleLogout = useCallback(() => dispatch(logout()), [dispatch])
  const user = useSelector(selectUser)

  const isAdmin = useSelector(selectUserIsAdmin)

  const dropdownItems = (
    <React.Fragment>
      <Dropdown.Item
        title={sharedMessages.profileSettings}
        icon={IconUserCircle}
        path="/console/user-settings/profile"
        external
      />
      {isAdmin && (
        <Dropdown.Item
          title={sharedMessages.adminPanel}
          icon={IconAdminPanel}
          path="/console/admin-panel/network-information"
          external
        />
      )}
      <hr />
      <Dropdown.Item
        title={sharedMessages.getSupport}
        icon={IconSupport}
        path="https://thethingsindustries.com/support"
        external
      />
      <Dropdown.Item
        title={sharedMessages.documentation}
        icon={IconBook}
        path="https://thethingsindustries.com/docs"
        external
      />
      <hr />
      <Dropdown.Item title={sharedMessages.logout} icon={IconLogout} action={handleLogout} />
    </React.Fragment>
  )

  const hasCustomBranding = selectBrandingRootPath() !== selectAssetsRootPath()
  const brandLogo = hasCustomBranding
    ? {
        src: `${selectBrandingRootPath()}/logo.svg`,
        alt: 'Logo',
      }
    : undefined

  return (
    <HeaderComponent
      user={user}
      profileDropdownItems={dropdownItems}
      brandLogo={brandLogo}
      Logo={Logo}
      onMenuClick={onMenuClick}
      alwaysShowLogo={alwaysShowLogo}
    />
  )
}

Header.propTypes = {
  alwaysShowLogo: PropTypes.bool,
  onMenuClick: PropTypes.func.isRequired,
}

Header.defaultProps = {
  alwaysShowLogo: false,
}

export default Header
