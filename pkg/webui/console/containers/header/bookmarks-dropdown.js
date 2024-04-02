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
import { defineMessages } from 'react-intl'

import Dropdown from '@ttn-lw/components/dropdown'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'
import useBookmark from '@ttn-lw/lib/hooks/use-bookmark'

import { selectBookmarksList } from '@console/store/selectors/user-preferences'

import style from './header.styl'

const m = defineMessages({
  noBookmarks: 'No bookmarks yet',
  noBookmarksDescription: 'Your bookmarked entities will be listed here',
  threshold: 'Only showing latest 15 bookmarks',
})

const Bookmark = ({ bookmark }) => {
  const { title, ids, path, icon } = useBookmark(bookmark)

  return (
    <Dropdown.Item
      title={title === '' ? ids.id : title}
      path={path}
      icon={icon}
      messageClassName={style.bookmark}
    />
  )
}

Bookmark.propTypes = {
  bookmark: PropTypes.shape({
    entity_ids: PropTypes.shape({}).isRequired,
  }).isRequired,
}

const BookmarksDropdown = () => {
  const dropdownItems = useSelector(selectBookmarksList)

  return dropdownItems && dropdownItems.length === 0 ? (
    <div className={style.emptyState}>
      <Message
        content={m.noBookmarks}
        className="d-block text-center fw-bold c-text-neutral-semilight"
      />
      <Message
        content={m.noBookmarksDescription}
        className="d-block text-center fs-s c-text-neutral-light"
      />
    </div>
  ) : (
    <>
      {dropdownItems.slice(0, 15).map(bookmark => (
        <Bookmark key={bookmark.created_at} bookmark={bookmark} />
      ))}
      {dropdownItems.length > 15 && (
        <div className="p-cs-l c-text-neutral-light fs-s text-center c-bg-brand-extralight br-l">
          <Message content={m.threshold} />
        </div>
      )}
    </>
  )
}

export default BookmarksDropdown
