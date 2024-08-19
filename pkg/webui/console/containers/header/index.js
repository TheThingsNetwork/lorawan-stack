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

import React, { useCallback, useContext } from 'react'
import { useDispatch, useSelector } from 'react-redux'

import { APPLICATION } from '@console/constants/entities'

import {
  IconLogout,
  IconUserCircle,
  IconBook,
  IconAdminPanel,
  IconApplication,
  IconDevice,
  IconGateway,
  IconOrganization,
  IconSupport,
} from '@ttn-lw/components/icon'
import HeaderComponent from '@ttn-lw/components/header'
import Dropdown from '@ttn-lw/components/dropdown'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import { selectAssetsRootPath, selectBrandingRootPath } from '@ttn-lw/lib/selectors/env'
import PropTypes from '@ttn-lw/lib/prop-types'

import {
  checkFromState,
  mayViewApplications,
  mayViewGateways,
  mayViewOrganizationsOfUser,
} from '@console/lib/feature-checks'

import { logout } from '@console/store/actions/logout'
import { setSearchOpen, setSearchScope } from '@console/store/actions/search'

import { selectUser, selectUserIsAdmin } from '@console/store/selectors/user'
import { selectTotalUnseenCount } from '@console/store/selectors/notifications'

import Logo from '../logo'
import SidebarContext from '../sidebar/context'

import NotificationsDropdown from './notifications-dropdown'
import BookmarksDropdown from './bookmarks-dropdown'

const Header = ({ alwaysShowLogo }) => {
  const dispatch = useDispatch()
  const {
    isMinimized,
    onMinimizeToggle,
    isDrawerOpen,
    openDrawer,
    closeDrawer,
    setIsHovered: setIsSideBarHovered,
  } = useContext(SidebarContext)

  const handleLogout = useCallback(() => dispatch(logout()), [dispatch])
  const user = useSelector(selectUser)
  const mayViewApps = useSelector(state =>
    user ? checkFromState(mayViewApplications, state) : false,
  )
  const mayViewGtws = useSelector(state => (user ? checkFromState(mayViewGateways, state) : false))
  const mayViewOrgs = useSelector(state =>
    user ? checkFromState(mayViewOrganizationsOfUser, state) : false,
  )
  const isAdmin = useSelector(selectUserIsAdmin)
  const hasUnseenNotifications = useSelector(selectTotalUnseenCount) > 0

  const onDrawerExpandClick = useCallback(() => {
    if (!isDrawerOpen) {
      openDrawer()
    } else {
      closeDrawer()
    }
  }, [isDrawerOpen, openDrawer, closeDrawer])

  const handleRegisterEndDeviceClick = useCallback(() => {
    dispatch(setSearchScope(APPLICATION))
    dispatch(setSearchOpen(true))
  }, [dispatch])

  const plusDropdownItems = (
    <>
      {mayViewApps && (
        <Dropdown.Item
          title={sharedMessages.addApplication}
          icon={IconApplication}
          path="/applications/add"
        />
      )}
      {mayViewGtws && (
        <Dropdown.Item title={sharedMessages.addGateway} icon={IconGateway} path="/gateways/add" />
      )}
      {mayViewOrgs && (
        <Dropdown.Item
          title={sharedMessages.addOrganization}
          icon={IconOrganization}
          path="/organizations/add"
        />
      )}

      <Dropdown.Item
        title={sharedMessages.registerDeviceInApplication}
        icon={IconDevice}
        action={handleRegisterEndDeviceClick}
      />
    </>
  )

  const dropdownItems = (
    <React.Fragment>
      <Dropdown.Item
        title={sharedMessages.profileSettings}
        icon={IconUserCircle}
        path="/user-settings/profile"
      />
      {isAdmin && (
        <Dropdown.Item
          title={sharedMessages.adminPanel}
          icon={IconAdminPanel}
          path="/admin-panel/network-information"
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

  const handleExpandSidebar = useCallback(() => {
    onDrawerExpandClick()
    setIsSideBarHovered(true)
  }, [onDrawerExpandClick, setIsSideBarHovered])

  const handleHideSidebar = useCallback(() => {
    setIsSideBarHovered(false)
  }, [setIsSideBarHovered])

  return (
    <HeaderComponent
      isSidebarMinimized={isMinimized}
      toggleSidebarMinimized={onMinimizeToggle}
      user={user}
      profileDropdownItems={dropdownItems}
      addDropdownItems={plusDropdownItems}
      bookmarkDropdownItems={<BookmarksDropdown />}
      notificationsDropdownItems={<NotificationsDropdown />}
      brandLogo={brandLogo}
      Logo={Logo}
      onMenuClick={onDrawerExpandClick}
      showNotificationDot={hasUnseenNotifications}
      alwaysShowLogo={alwaysShowLogo}
      expandSidebar={handleExpandSidebar}
      handleHideSidebar={handleHideSidebar}
    />
  )
}

Header.propTypes = {
  alwaysShowLogo: PropTypes.bool,
}

Header.defaultProps = {
  alwaysShowLogo: false,
}

export default Header
