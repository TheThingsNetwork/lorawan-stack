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

import SideNavigation from '@ttn-lw/components/sidebar/side-menu'

import useBookmark from '@ttn-lw/lib/hooks/use-bookmark'
import PropTypes from '@ttn-lw/lib/prop-types'

import { selectGatewayBookmarks } from '@console/store/selectors/user-preferences'
import { selectUserId } from '@console/store/selectors/logout'

import SidebarContext from '../context'

import TopEntitiesSection from './top-entities-section'

const Bookmark = ({ bookmark }) => {
  const { title, ids, path, icon } = useBookmark(bookmark)

  return <SideNavigation.Item title={title === '' ? ids.id : title} path={path} icon={icon} />
}

Bookmark.propTypes = {
  bookmark: PropTypes.shape({}).isRequired,
}

const GtwListSideNavigation = () => {
  const topEntities = useSelector(state => selectGatewayBookmarks(state))
  const { isMinimized } = useContext(SidebarContext)
  const userId = useSelector(selectUserId)

  if (isMinimized || topEntities.length === 0) {
    return <div />
  }

  return <TopEntitiesSection topEntities={topEntities} userId={userId} />
}

export default GtwListSideNavigation
