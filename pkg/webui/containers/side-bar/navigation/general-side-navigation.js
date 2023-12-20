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

import SideNavigation from '@ttn-lw/components/navigation/side-v2'
import SectionLabel from '@ttn-lw/components/sidebar/section-label'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import {
  checkFromState,
  mayViewOrEditApiKeys,
  mayViewOrganizationsOfUser,
} from '@console/lib/feature-checks'

import { selectUser, selectUserIsAdmin } from '@console/store/selectors/logout'

import SidebarContext from '../context'

const GeneralSideNavigation = () => {
  const { topEntities, isMinimized } = useContext(SidebarContext)

  const isUserAdmin = useSelector(selectUserIsAdmin)
  const user = useSelector(selectUser)
  const mayViewOrgs = useSelector(state =>
    user ? checkFromState(mayViewOrganizationsOfUser, state) : false,
  )
  const mayHandleApiKeys = useSelector(state =>
    user ? checkFromState(mayViewOrEditApiKeys, state) : false,
  )

  return (
    <div>
      <SideNavigation className="mt-cs-xs">
        <SideNavigation.Item title={sharedMessages.dashboard} path="/" icon="overview" exact />
        {mayViewOrgs && (
          <SideNavigation.Item
            title={sharedMessages.organizations}
            path="/organizations"
            icon="group"
          />
        )}
        <SideNavigation.Item
          title={sharedMessages.notifications}
          path="/notifications"
          icon="inbox"
        />
        {mayHandleApiKeys && (
          <SideNavigation.Item
            title={sharedMessages.personalApiKeys}
            path="/user/api-keys"
            icon="key"
          />
        )}
        {isUserAdmin && (
          <SideNavigation.Item
            title={sharedMessages.adminPanel}
            path="/admin-panel"
            icon="admin_panel_settings"
          />
        )}
        {!isMinimized && (
          <>
            <SectionLabel label="Top entities" icon="add" className="mt-cs-m" />
            {topEntities.map(({ path, title, entity }) => (
              <SideNavigation.Item key={path} title={title} path={path} icon={entity} />
            ))}
          </>
        )}
      </SideNavigation>
    </div>
  )
}

export default GeneralSideNavigation
