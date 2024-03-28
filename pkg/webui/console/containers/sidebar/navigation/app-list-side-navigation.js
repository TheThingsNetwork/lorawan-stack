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

import { IconPlus } from '@ttn-lw/components/icon'
import SectionLabel from '@ttn-lw/components/sidebar/section-label'
import SideNavigation from '@ttn-lw/components/sidebar/side-menu'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import SidebarContext from '../context'

const AppListSideNavigation = () => {
  const { topEntities, isMinimized } = useContext(SidebarContext)

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
        {topEntities.map(({ path, entity, title }) => (
          <SideNavigation.Item title={title} path={path} icon={entity} key={path} />
        ))}
      </SideNavigation>
    </div>
  )
}

export default AppListSideNavigation