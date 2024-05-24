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

import { selectUserId } from '@console/store/selectors/logout'
import { selectPerEntityBookmarks } from '@console/store/selectors/user-preferences'

import SidebarContext from '../context'

import TopEntitiesSection from './top-entities-section'

const AppListSideNavigation = () => {
  const topEntities = useSelector(selectPerEntityBookmarks('application'))
  const { isMinimized } = useContext(SidebarContext)
  const userId = useSelector(selectUserId)

  if (isMinimized || topEntities.length === 0) {
    // Rendering an empty div to prevent the shadow of the search bar
    // from being cut off. There will be a default element rendering
    // here in the future anyway.
    return <div />
  }

  return <TopEntitiesSection topEntities={topEntities} userId={userId} entity="application" />
}

export default AppListSideNavigation
