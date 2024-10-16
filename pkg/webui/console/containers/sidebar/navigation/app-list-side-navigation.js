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

import { APPLICATION } from '@console/constants/entities'

import { selectApplicationTopEntities } from '@console/store/selectors/top-entities'

import TopEntitiesSection from './top-entities-section'

const AppListSideNavigation = () => {
  const topEntities = useSelector(selectApplicationTopEntities)

  if (topEntities.length === 0) {
    // Rendering an empty div to prevent the shadow of the search bar
    // from being cut off. There will be a default element rendering
    // here in the future anyway.
    return <div />
  }

  return <TopEntitiesSection topEntities={topEntities} type={APPLICATION} />
}

export default AppListSideNavigation
