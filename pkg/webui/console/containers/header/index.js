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
import NavigationBar from '@ttn-lw/components/navigation/bar'
import Dropdown from '@ttn-lw/components/dropdown'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

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

const Header = ({ searchable, handleSearchRequest }) => {
  const dispatch = useDispatch()

  const handleLogout = useCallback(() => dispatch(logout()), [dispatch])
  const user = useSelector(selectUser)
  const isUserAdmin = useSelector(selectUserIsAdmin)
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

  const navigation = [
    {
      title: sharedMessages.overview,
      icon: 'overview',
      path: '',
      exact: true,
      hidden: !mayViewApps && !mayViewGateways,
    },
    {
      title: sharedMessages.applications,
      icon: 'application',
      path: '/applications',
      hidden: !mayViewApps,
    },
    {
      title: sharedMessages.gateways,
      icon: 'gateway',
      path: '/gateways',
      hidden: !mayViewGtws,
    },
    {
      title: sharedMessages.organizations,
      icon: 'organization',
      path: '/organizations',
      hidden: !mayViewOrgs,
    },
  ]

  const navigationEntries = (
    <React.Fragment>
      {navigation.map(
        ({ hidden, ...rest }) => !hidden && <NavigationBar.Item {...rest} key={rest.title.id} />,
      )}
    </React.Fragment>
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

  const mobileDropdownItems = (
    <React.Fragment>
      {navigation.map(
        ({ hidden, ...rest }) => !hidden && <Dropdown.Item {...rest} key={rest.title.id} />,
      )}
      <React.Fragment>
        <hr />
        <Dropdown.Item
          title={sharedMessages.profileSettings}
          icon="user"
          path={`${accountUrl}/profile-settings`}
          external
        />
      </React.Fragment>
      {mayHandleApiKeys && (
        <Dropdown.Item
          title={sharedMessages.personalApiKeys}
          icon="api_keys"
          path="/user/api-keys"
        />
      )}
      {isUserAdmin && (
        <Dropdown.Item
          title={sharedMessages.adminPanel}
          icon="lock"
          path="/admin-panel/network-information"
        />
      )}
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
    </React.Fragment>
  )

  return (
    <HeaderComponent
      user={user}
      dropdownItems={dropdownItems}
      mobileDropdownItems={mobileDropdownItems}
      navigationEntries={navigationEntries}
      searchable={searchable}
      onSearchRequest={handleSearchRequest}
      onLogout={handleLogout}
      logo={<Logo />}
    />
  )
}

Header.propTypes = {
  /** A handler for when the user used the search input. */
  handleSearchRequest: PropTypes.func,
  /** A flag identifying whether the header should display the search input. */
  searchable: PropTypes.bool,
}

Header.defaultProps = {
  handleSearchRequest: () => null,
  searchable: false,
}

export default Header
