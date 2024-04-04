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

import React, { useCallback, useContext, useState } from 'react'
import { useSelector } from 'react-redux'

import {
  IconUsersGroup,
  IconLayoutDashboard,
  IconUserShield,
  IconKey,
  IconPlus,
  IconInbox,
} from '@ttn-lw/components/icon'
import SideNavigation from '@ttn-lw/components/sidebar/side-menu'
import SectionLabel from '@ttn-lw/components/sidebar/section-label'
import Button from '@ttn-lw/components/button'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'
import useBookmark from '@ttn-lw/lib/hooks/use-bookmark'

import {
  checkFromState,
  mayViewOrEditApiKeys,
  mayViewOrganizationsOfUser,
} from '@console/lib/feature-checks'

import { selectUser, selectUserIsAdmin } from '@console/store/selectors/logout'
import { selectBookmarksList } from '@console/store/selectors/user-preferences'

import SidebarContext from '../context'

const Bookmark = ({ bookmark }) => {
  const { title, ids, path, icon } = useBookmark(bookmark)

  return <SideNavigation.Item title={title === '' ? ids.id : title} path={path} icon={icon} />
}

Bookmark.propTypes = {
  bookmark: PropTypes.shape({}).isRequired,
}

const GeneralSideNavigation = () => {
  const { isMinimized } = useContext(SidebarContext)
  const [showMore, setShowMore] = useState(false)
  const topEntities = useSelector(state => selectBookmarksList(state))
  const isUserAdmin = useSelector(selectUserIsAdmin)
  const user = useSelector(selectUser)
  const mayViewOrgs = useSelector(state =>
    user ? checkFromState(mayViewOrganizationsOfUser, state) : false,
  )
  const mayHandleApiKeys = useSelector(state =>
    user ? checkFromState(mayViewOrEditApiKeys, state) : false,
  )

  const handleShowMore = useCallback(async () => {
    setShowMore(showMore => !showMore)
  }, [])

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
            path="/organizations"
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
            path="/user/api-keys"
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
      {!isMinimized && topEntities.length > 0 && (
        <SideNavigation className="mt-cs-xs">
          <SectionLabel label={sharedMessages.topEntities} icon={IconPlus} onClick={() => null} />
          {topEntities.slice(0, 6).map(bookmark => (
            <Bookmark key={bookmark.created_at} bookmark={bookmark} />
          ))}
          {showMore &&
            topEntities.length > 6 &&
            topEntities
              .slice(6, topEntities.length)
              .map(bookmark => <Bookmark key={bookmark.created_at} bookmark={bookmark} />)}
          {topEntities.length > 6 && (
            <Button
              message={showMore ? sharedMessages.showLess : sharedMessages.showMore}
              onClick={handleShowMore}
              className="c-text-neutral-light ml-cs-xs mt-cs-xs fs-s"
            />
          )}
        </SideNavigation>
      )}
    </>
  )
}

export default GeneralSideNavigation
