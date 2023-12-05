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

import SideNavigation from '@ttn-lw/components/navigation/side-v2'
import SectionLabel from '@ttn-lw/components/section-label'

import SideBarContext from '../context'

const GeneralSideNavigation = () => {
  const { topEntities } = useContext(SideBarContext)

  return (
    <div>
      <SideNavigation>
        <SideNavigation.Item title="Dashboard" path="/" icon="overview" exact />
        <SideNavigation.Item title="Organizations" path="/organizations" icon="group" />
        <SideNavigation.Item title="Notifications" path="/notifications" icon="notifications" />
        <SideNavigation.Item title="Personal API keys" path="/user/api-keys" icon="key" />
        <SideNavigation.Item title="Network Operations Center" path="/noc" icon="ssid_chart" />
        <SideNavigation.Item title="Admin panel" path="/admin-panel" icon="admin_panel_settings" />
        <SectionLabel label="Top entities" icon="add" className="mt-cs-m" />
        {topEntities.map(({ path, title, entity }) => (
          <SideNavigation.Item key={path} title={title} path={path} icon={entity} />
        ))}
      </SideNavigation>
    </div>
  )
}

export default GeneralSideNavigation
