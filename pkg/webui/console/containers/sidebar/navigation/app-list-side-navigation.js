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

import { IconPlus } from '@ttn-lw/components/icon'
import SectionLabel from '@ttn-lw/components/sidebar/section-label'
import SideNavigation from '@ttn-lw/components/sidebar/side-menu'
import Button from '@ttn-lw/components/button'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'
import useBookmark from '@ttn-lw/lib/hooks/use-bookmark'

import { selectApplicationBookmarks } from '@console/store/selectors/user-preferences'

import SidebarContext from '../context'

const Bookmark = ({ bookmark }) => {
  const { title, ids, path, icon } = useBookmark(bookmark)

  return <SideNavigation.Item title={title === '' ? ids.id : title} path={path} icon={icon} />
}

Bookmark.propTypes = {
  bookmark: PropTypes.shape({}).isRequired,
}

const AppListSideNavigation = () => {
  const [showMore, setShowMore] = useState(false)
  const topEntities = useSelector(state => selectApplicationBookmarks(state))
  const { isMinimized } = useContext(SidebarContext)

  const handleShowMore = useCallback(async () => {
    setShowMore(showMore => !showMore)
  }, [])

  if (isMinimized || topEntities.length === 0) {
    // Rendering an empty div to prevent the shadow of the search bar
    // from being cut off. There will be a default element rendering
    // here in the future anyway.
    return <div />
  }

  return (
    <div>
      <SectionLabel label={sharedMessages.topApplications} icon={IconPlus} onClick={() => null} />
      <SideNavigation>
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
    </div>
  )
}

export default AppListSideNavigation
