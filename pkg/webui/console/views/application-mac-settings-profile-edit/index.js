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

import React from 'react'
import { useParams } from 'react-router-dom'
import { useSelector } from 'react-redux'

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import PageTitle from '@ttn-lw/components/page-title'

import RequireRequest from '@ttn-lw/lib/components/require-request'

import MACSettingsProfileForm from '@console/containers/mac-settings-profile-form'

import { getMacSettingsProfile } from '@console/store/actions/mac-settings-profiles'

import { selectMacSettingsProfileById } from '@console/store/selectors/mac-settings-profiles'

const ApplicationMacSettingsProfilesEdit = () => {
  const { appId, macSettingsProfileId } = useParams()
  const macSettingsProfile = useSelector(state =>
    selectMacSettingsProfileById(state, macSettingsProfileId),
  )

  useBreadcrumbs(
    'apps.single.mac-settings-profiles.single',
    <Breadcrumb
      path={`/applications/${appId}/mac-settings-profiles/${macSettingsProfileId}`}
      content={macSettingsProfileId}
    />,
  )

  return (
    <RequireRequest requestAction={getMacSettingsProfile(appId, macSettingsProfileId)}>
      <div className="container container--lg grid">
        <PageTitle title={'Edit a MAC settings profile'} className="mb-0" />
        <div className="item-12">
          <MACSettingsProfileForm
            edit
            macSettingsProfile={macSettingsProfile?.mac_settings_profile}
            macSettingsProfileId={macSettingsProfileId}
          />
        </div>
      </div>
    </RequireRequest>
  )
}

export default ApplicationMacSettingsProfilesEdit
