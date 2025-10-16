// Copyright © 2025 The Things Network Foundation, The Things Industries B.V.
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

import React, { useCallback } from 'react'
import { useParams } from 'react-router-dom'
import { createSelector } from 'reselect'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'

import MacSettingsProfilesTable from '@console/containers/mac-settings-profiles-table'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { getMacSettingsProfilesList } from '@console/store/actions/mac-settings-profiles'

import {
  selectMacSettingsProfiles,
  selectMacSettingsProfilesTotalCount,
} from '@console/store/selectors/mac-settings-profiles'

const ApplicationMacSettingsProfilesList = () => {
  const { appId } = useParams()

  const getMacSettingsProfiles = useCallback(
    filters => getMacSettingsProfilesList(undefined, appId, filters),
    [appId],
  )

  const baseDataSelector = createSelector(
    [selectMacSettingsProfiles, selectMacSettingsProfilesTotalCount],
    (macSettingsProfiles, totalCount) => ({
      mac_settings_profiles: macSettingsProfiles,
      totalCount,
    }),
  )

  return (
    <div className="container container--xl">
      <IntlHelmet title={sharedMessages.macSettingsProfiles} />
      <MacSettingsProfilesTable
        baseDataSelector={baseDataSelector}
        getItemsAction={getMacSettingsProfiles}
      />
    </div>
  )
}

export default ApplicationMacSettingsProfilesList
