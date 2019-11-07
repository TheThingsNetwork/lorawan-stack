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

import { logout } from '../../store/actions/user'
import { selectUser } from '../../store/selectors/user'
import {
  checkFromState,
  mayViewApplications,
  mayViewGateways,
  mayViewOrganizationsOfUser,
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
      }
    }
    return { user }
  },
  { handleLogout: logout },
)
@bind
class Header extends Component {
  static propTypes = {
    /**
     * The User object, retrieved from the API. If it is `undefined`, then the
     * guest header is rendered
     */
    anchored: PropTypes.bool,
    /** Flag identifying whether links should be rendered as plain anchor link */
    handleLogout: PropTypes.func.isRequired,
    /** A handler for when the user used the search input */
    handleSearchRequest: PropTypes.func,
    /** A handler for when the user clicks the logout button */
    searchable: PropTypes.bool,
    /** A flag identifying whether the header should display the search input */
    user: PropTypes.object,
    /** The rights of the current user */
    rights: PropTypes.rights,
  }

  render() {
    const {
      user,
      anchored,
      handleSearchRequest,
      handleLogout,
      searchable,
      mayViewApplications,
      mayViewGateways,
      mayViewOrganizations,
    } = this.props

    const navigationEntries = [
      {
        title: sharedMessages.overview,
        icon: 'overview',
        path: '/',
        exact: true,
      },
      {
        title: sharedMessages.applications,
        icon: 'application',
        path: '/applications',
        hidden: !mayViewApplications,
      },
      {
        title: sharedMessages.gateways,
        icon: 'gateway',
        path: '/gateways',
        hidden: !mayViewGateways,
      },
      {
        title: sharedMessages.organizations,
        icon: 'organization',
        path: '/organizations',
        hidden: !mayViewOrganizations,
      },
    ]

    const dropdownItems = [
      {
        title: sharedMessages.logout,
        icon: 'power_settings_new',
        action: handleLogout,
      },
    ]

    return (
      <HeaderComponent
        user={user}
        dropdownItems={dropdownItems}
        navigationEntries={navigationEntries}
        anchored={anchored}
        searchable={searchable}
        handleSearchRequest={handleSearchRequest}
      />
    )
  }
}

export default Header
