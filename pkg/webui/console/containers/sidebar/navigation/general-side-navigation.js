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

import React, { useContext } from 'react'
import { useSelector } from 'react-redux'

import { PAGE_SIZES } from '@ttn-lw/constants/page-sizes'

import {
  IconUsersGroup,
  IconLayoutDashboard,
  IconUserShield,
  IconKey,
  IconInbox,
} from '@ttn-lw/components/icon'
import SideNavigation from '@ttn-lw/components/sidebar/side-menu'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import {
  checkFromState,
  mayViewOrEditApiKeys,
  mayViewOrganizationsOfUser,
} from '@console/lib/feature-checks'
import getCookie from '@console/lib/table-utils'

import { selectUser, selectUserIsAdmin } from '@console/store/selectors/logout'
import { selectBookmarksList } from '@console/store/selectors/user-preferences'

import SidebarContext from '../context'

import TopEntitiesSection from './top-entities-section'

const GeneralSideNavigation = () => {
  const { isMinimized } = useContext(SidebarContext)
  const topEntities = useSelector(state => selectBookmarksList(state))
  const isUserAdmin = useSelector(selectUserIsAdmin)
  const user = useSelector(selectUser)
  const mayViewOrgs = useSelector(state =>
    user ? checkFromState(mayViewOrganizationsOfUser, state) : false,
  )
  const mayHandleApiKeys = useSelector(state =>
    user ? checkFromState(mayViewOrEditApiKeys, state) : false,
  )

  const orgPageSize = getCookie('organizations-list-page-size')
  const orgParam = `?page-size=${orgPageSize ? orgPageSize : PAGE_SIZES.REGULAR}`
  const keysPageSize = getCookie('keys-list-page-size')
  const keysParam = `?page-size=${keysPageSize ? keysPageSize : PAGE_SIZES.REGULAR}`

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
            path={`/organizations${orgParam}`}
            icon={IconUsersGroup}
          />
        )}
        <SideNavigation.Item
          title={sharedMessages.notifications}
          path="/notifications"
          icon={IconInbox}
        />
        {mayHandleApiKeys && (
          <SideNavigation.Item
            title={sharedMessages.personalApiKeys}
            path={`/user/api-keys${keysParam}`}
            icon={IconKey}
          />
        )}
        {isUserAdmin && (
          <SideNavigation.Item
            title={sharedMessages.adminPanel}
            path="/admin-panel"
            icon={IconUserShield}
          />
        )}
      </SideNavigation>
      {!isMinimized && <TopEntitiesSection topEntities={topEntities} userId={user.ids.user_id} />}
    </>
  )
}

export default GeneralSideNavigation
