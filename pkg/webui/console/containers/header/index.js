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

import React, { Component } from 'react'
import { connect } from 'react-redux'
import { withRouter } from 'react-router-dom'
import bind from 'autobind-decorator'

import PropTypes from '../../../lib/prop-types'
import sharedMessages from '../../../lib/shared-messages'
import HeaderComponent from '../../../components/header'
import NavigationBar from '../../../components/navigation/bar'
import Dropdown from '../../../components/dropdown'

import { logout } from '../../store/actions/user'
import { selectUser } from '../../store/selectors/user'
import {
  checkFromState,
  mayViewApplications,
  mayViewGateways,
  mayViewOrganizationsOfUser,
  mayManageUsers,
} from '../../lib/feature-checks'

@withRouter
@connect(
  function(state) {
    const user = selectUser(state)
    if (Boolean(user)) {
      return {
        user,
        mayViewApplications: checkFromState(mayViewApplications, state),
        mayViewGateways: checkFromState(mayViewGateways, state),
        mayViewOrganizations: checkFromState(mayViewOrganizationsOfUser, state),
        mayManageUsers: checkFromState(mayManageUsers, state),
      }
    }
    return { user }
  },
  { handleLogout: logout },
)
@bind
class Header extends Component {
  static propTypes = {
    /** A handler for when the user clicks the logout button */
    handleLogout: PropTypes.func.isRequired,
    /** A handler for when the user used the search input */
    handleSearchRequest: PropTypes.func,
    mayManageUsers: PropTypes.bool,
    mayViewApplications: PropTypes.bool,
    mayViewGateways: PropTypes.bool,
    mayViewOrganizations: PropTypes.bool,
    /** A flag identifying whether the header should display the search input */
    searchable: PropTypes.bool,
    /**
     * The User object, retrieved from the API. If it is `undefined`, then the
     * guest header is rendered
     */
    user: PropTypes.user,
  }

  static defaultProps = {
    handleSearchRequest: () => null,
    searchable: false,
    user: undefined,
    mayManageUsers: false,
    mayViewApplications: false,
    mayViewGateways: false,
    mayViewOrganizations: false,
  }

  render() {
    const {
      user,
      handleSearchRequest,
      handleLogout,
      searchable,
      mayViewApplications,
      mayViewGateways,
      mayViewOrganizations,
      mayManageUsers,
    } = this.props

    const navigationEntries = (
      <React.Fragment>
        <NavigationBar.Item title={sharedMessages.overview} icon="overview" path="/" exact />
        {mayViewApplications && (
          <NavigationBar.Item
            title={sharedMessages.applications}
            icon="application"
            path="/applications"
          />
        )}
        {mayViewGateways && (
          <NavigationBar.Item title={sharedMessages.gateways} icon="gateway" path="/gateways" />
        )}
        {mayViewOrganizations && (
          <NavigationBar.Item
            title={sharedMessages.organizations}
            icon="organization"
            path="/organizations"
          />
        )}
      </React.Fragment>
    )

    const dropdownItems = (
      <React.Fragment>
        {mayManageUsers && (
          <Dropdown.Item
            title={sharedMessages.userManagement}
            icon="user_management"
            path="/admin/user-management"
          />
        )}
        <Dropdown.Item
          title={sharedMessages.logout}
          icon="power_settings_new"
          action={handleLogout}
        />
      </React.Fragment>
    )

    return (
      <HeaderComponent
        user={user}
        dropdownItems={dropdownItems}
        navigationEntries={navigationEntries}
        searchable={searchable}
        handleSearchRequest={handleSearchRequest}
      />
    )
  }
}

export default Header
