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

import React, { useEffect } from 'react'

import SectionLabel from '@ttn-lw/components/section-label'
import SideNavigation from '@ttn-lw/components/navigation/side-v2'

import getCookie from '../utils'

const AppListSideNavigation = () => {
/*   // To be removed
  const cookie = [
    JSON.stringify({
      path: '/applications/parking-lot-occupancy',
      entity: 'application',
      title: 'Parking Lot Occupancy',
      tag: 'general',
    }),
    JSON.stringify({
      path: '/applications/parking-lot-occupancy',
      entity: 'application',
      title: 'Parking Lot Occupancy',
      tag: 'application',
    }),
    JSON.stringify({
      path: '/applications/parking-lot-occupancy',
      entity: 'gateway',
      title: 'Parking Lot Occupancy',
      tag: 'gateway',
    }),
    JSON.stringify({
      path: '/applications/parking-lot-occupancy',
      entity: 'device',
      title: 'Parking Lot Occupancy',
      tag: 'device',
    }),
    JSON.stringify({
      path: '/applications/parking-lot-occupancy',
      entity: 'device',
      title: 'Parking Lot Occupancy',
      tag: 'general',
    }),
  ].join('_')

  useEffect(() => {
    document.cookie = `topEntities=${cookie};expires=1365094617464;domain=localhost;SameSite=None`
  }, [cookie])
  // End of To be removed */

  const topEntitiesCookie = getCookie('topEntities')
    .split('_')
    .map(cookie => JSON.parse(cookie))

  const topEntities = topEntitiesCookie.filter(cookie => cookie.tag === 'application')

  return (
    <div>
      <SectionLabel label="Top entities" icon="add" />
      <SideNavigation>
        {topEntities.map(({ path, entity, title }) => (
          <SideNavigation.Item title={title} path={path} icon={entity} key={path} />
        ))}
      </SideNavigation>
    </div>
  )
}

export default AppListSideNavigation
