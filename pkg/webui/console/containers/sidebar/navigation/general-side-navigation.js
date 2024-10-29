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

import React from 'react'
import { useSelector } from 'react-redux'

import {
  IconUsersGroup,
  IconLayoutDashboard,
  IconUserShield,
  IconInbox,
  IconAperture,
  IconTextCaption,
  IconUserCircle,
  IconPassword,
  IconUserCog,
  IconApiKeys,
  IconShieldLock,
  IconLockOpen,
  IconMailCog,
} from '@ttn-lw/components/icon'
import SideNavigation from '@ttn-lw/components/sidebar/side-menu'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import {
  checkFromState,
  mayConfigurePacketBroker,
  mayManageUsers,
  mayViewClientsOfUser,
  mayViewOrEditClientAuthorizations,
  mayViewOrEditUserApiKeys,
  mayViewOrEditUserSettings,
  mayViewOrganizationsOfUser,
} from '@console/lib/feature-checks'

import { selectUser, selectUserIsAdmin } from '@console/store/selectors/user'
import { selectTopEntitiesAll } from '@console/store/selectors/top-entities'

import TopEntitiesSection from './top-entities-section'

const GeneralSideNavigation = () => {
  const topEntities = useSelector(selectTopEntitiesAll)
  const isUserAdmin = useSelector(selectUserIsAdmin)
  const user = useSelector(selectUser)
  const mayViewOrgs = useSelector(state =>
    user ? checkFromState(mayViewOrganizationsOfUser, state) : false,
  )
  const showUserManagement = useSelector(state => checkFromState(mayManageUsers, state))
  const showPacketBroker = useSelector(state => checkFromState(mayConfigurePacketBroker, state))
  const showProfileSettings = useSelector(state => checkFromState(mayViewOrEditUserSettings, state))
  const showUserApiKeys = useSelector(state => checkFromState(mayViewOrEditUserApiKeys, state))
  const showAuthorizationManagement = useSelector(state =>
    checkFromState(mayViewOrEditClientAuthorizations, state),
  )
  const showClientManagement = useSelector(state => checkFromState(mayViewClientsOfUser, state))

  return (
    <>
      <SideNavigation>
        <SideNavigation.Item
          title={sharedMessages.dashboard}
          path="/"
          icon={IconLayoutDashboard}
          exact
        />
        {mayViewOrgs && (
          <SideNavigation.Item
            title={sharedMessages.organizations}
            path={`/organizations`}
            icon={IconUsersGroup}
          />
        )}
        <SideNavigation.Item
          title={sharedMessages.notifications}
          path="/notifications/inbox"
          icon={IconInbox}
        />
        {isUserAdmin && (
          <SideNavigation.Item
            title={sharedMessages.adminPanel}
            path="/admin-panel"
            icon={IconUserShield}
          >
            <SideNavigation.Item
              title={sharedMessages.networkInformation}
              path="/admin-panel/network-information"
              icon={IconTextCaption}
            />
            {showUserManagement && (
              <SideNavigation.Item
                title={sharedMessages.userManagement}
                path="/admin-panel/user-management"
                icon={IconUsersGroup}
              />
            )}
            {showPacketBroker && (
              <SideNavigation.Item
                title={sharedMessages.packetBroker}
                path="/admin-panel/packet-broker"
                icon={IconAperture}
              />
            )}
          </SideNavigation.Item>
        )}
        <SideNavigation.Item title={sharedMessages.userSettings} icon={IconUserCog}>
          {showProfileSettings && (
            <SideNavigation.Item
              title={sharedMessages.profile}
              path="/user-settings/profile"
              icon={IconUserCircle}
            />
          )}
          {showProfileSettings && (
            <SideNavigation.Item
              title={sharedMessages.password}
              path="/user-settings/password"
              icon={IconPassword}
            />
          )}
          {showProfileSettings && (
            <SideNavigation.Item
              title={'Email notifications'}
              path="/user-settings/email-notifications-settings"
              icon={IconMailCog}
            />
          )}
          {showUserApiKeys && (
            <SideNavigation.Item
              title={sharedMessages.apiKeys}
              path="/user-settings/api-keys"
              icon={IconApiKeys}
            />
          )}
          <SideNavigation.Item
            title={sharedMessages.sessionManagement}
            path="/user-settings/sessions"
            icon={IconShieldLock}
          />
          {showAuthorizationManagement && (
            <SideNavigation.Item
              title={sharedMessages.authorizations}
              path="/user-settings/authorizations"
              icon={IconLockOpen}
            />
          )}
          {showClientManagement && (
            <SideNavigation.Item
              title={sharedMessages.oauthClients}
              path="/user-settings/oauth-clients"
              icon={IconUserShield}
            />
          )}
        </SideNavigation.Item>
      </SideNavigation>
      <TopEntitiesSection topEntities={topEntities} />
    </>
  )
}

export default GeneralSideNavigation
