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

import React, { useCallback, useState } from 'react'

import { IconPlus } from '@ttn-lw/components/icon'
import Button from '@ttn-lw/components/button'
import SideNavigation from '@ttn-lw/components/sidebar/side-menu'
import SectionLabel from '@ttn-lw/components/sidebar/section-label'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'
import useBookmark from '@ttn-lw/lib/hooks/use-bookmark'

const Bookmark = ({ bookmark }) => {
  const { title, ids, path, icon } = useBookmark(bookmark)

  return <SideNavigation.Item title={title === '' ? ids.id : title} path={path} icon={icon} />
}

Bookmark.propTypes = {
  bookmark: PropTypes.shape({}).isRequired,
}

const TopEntitiesSection = ({ topEntities, entity }) => {
  const [showMore, setShowMore] = useState(false)

  const handleShowMore = useCallback(async () => {
    setShowMore(showMore => !showMore)
  }, [])

  const label = entity
    ? entity === 'gateway'
      ? sharedMessages.topGateways
      : sharedMessages.topApplications
    : sharedMessages.topEntities

  return (
    <SideNavigation>
      <SectionLabel label={label} icon={IconPlus} onClick={() => null} />
      {topEntities.slice(0, 6).map((bookmark, index) => (
        <Bookmark key={index} bookmark={bookmark} />
      ))}
      {showMore &&
        topEntities.length > 6 &&
        topEntities
          .slice(6, topEntities.length)
          .map((bookmark, index) => <Bookmark key={index} bookmark={bookmark} />)}
      {topEntities.length > 6 && (
        <Button
          message={showMore ? sharedMessages.showLess : sharedMessages.showMore}
          onClick={handleShowMore}
          className="c-text-neutral-light ml-cs-xs mt-cs-xs fs-s"
        />
      )}
    </SideNavigation>
  )
}

TopEntitiesSection.propTypes = {
  entity: PropTypes.string,
  topEntities: PropTypes.arrayOf(PropTypes.shape({})).isRequired,
}

TopEntitiesSection.defaultProps = {
  entity: undefined,
}

export default TopEntitiesSection
