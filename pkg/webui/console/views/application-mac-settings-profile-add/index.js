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

import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import PageTitle from '@ttn-lw/components/page-title'

import MACSettingsProfileForm from '@console/containers/mac-settings-profile-form'

import sharedMessages from '@ttn-lw/lib/shared-messages'

const ApplicationMacSettingsProfilesAdd = () => {
  const { appId } = useParams()
  useBreadcrumbs(
    'apps.single.mac-settings-profiles.add',
    <Breadcrumb
      path={`/applications/${appId}/mac-settings-profiles/add`}
      content={sharedMessages.add}
    />,
  )

  return (
    <div className="container container--lg grid">
      <PageTitle title={sharedMessages.createMacSettingsProfile} className="mb-0" />
      <div className="item-12">
        <MACSettingsProfileForm />
      </div>
    </div>
  )
}

export default ApplicationMacSettingsProfilesAdd
