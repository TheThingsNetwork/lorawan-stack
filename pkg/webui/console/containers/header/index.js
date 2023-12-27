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

import HeaderComponent from '@ttn-lw/components/header'
import Dropdown from '@ttn-lw/components/dropdown'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import { selectAssetsRootPath, selectBrandingRootPath } from '@ttn-lw/lib/selectors/env'

import selectAccountUrl from '@console/lib/selectors/app-config'
import {
  checkFromState,
  mayViewApplications,
  mayViewGateways,
  mayViewOrganizationsOfUser,
  mayViewOrEditApiKeys,
} from '@console/lib/feature-checks'

import { logout } from '@console/store/actions/logout'

import { selectUser, selectUserIsAdmin } from '@console/store/selectors/logout'

import Logo from '../logo'

const accountUrl = selectAccountUrl()

const Header = () => {
  const dispatch = useDispatch()

  const handleLogout = useCallback(() => dispatch(logout()), [dispatch])
  const user = useSelector(selectUser)
  const mayViewApps = useSelector(state =>
    user ? checkFromState(mayViewApplications, state) : false,
  )
  const mayViewGtws = useSelector(state => (user ? checkFromState(mayViewGateways, state) : false))
  const mayViewOrgs = useSelector(state =>
    user ? checkFromState(mayViewOrganizationsOfUser, state) : false,
  )
  const mayHandleApiKeys = useSelector(state =>
    user ? checkFromState(mayViewOrEditApiKeys, state) : false,
  )

  const plusDropdownItems = (
    <>
      {mayViewApps && (
        <Dropdown.Item
          title="Add new application"
          icon="display_settings"
          path="/applications/add"
        />
      )}
      {mayViewGtws && <Dropdown.Item title="Add new gateway" icon="router" path="/gateways/add" />}
      {mayViewOrgs && (
        <Dropdown.Item title="Add new organization" icon="support" path="/organizations/add" />
      )}

      <Dropdown.Item
        title="Register end device in application"
        icon="settings_remote"
        path="/devices/add"
      />
    </>
  )

  const dropdownItems = (
    <React.Fragment>
      <Dropdown.Item
        title={sharedMessages.profileSettings}
        icon="user"
        path={`${accountUrl}/profile-settings`}
        external
      />
      {mayHandleApiKeys && (
        <Dropdown.Item title={sharedMessages.apiKeys} icon="api_keys" path="/user/api-keys" />
      )}
      <Dropdown.Item
        title={sharedMessages.adminPanel}
        icon="lock"
        path="/admin-panel/network-information"
      />
      <hr />
      <Dropdown.Item
        title={sharedMessages.getSupport}
        icon="help"
        path="https://thethingsindustries.com/support"
        external
      />
      <Dropdown.Item
        title={sharedMessages.documentation}
        icon="description"
        path="https://thethingsindustries.com/docs"
        external
      />
      <hr />
      <Dropdown.Item title={sharedMessages.logout} icon="logout" action={handleLogout} />
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
      addDropdownItems={plusDropdownItems}
      starDropdownItems={[]}
      brandLogo={brandLogo}
      logo={<Logo />}
    />
  )
}

export default Header
