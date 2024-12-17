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
import { defineMessages } from 'react-intl'

import PageTitle from '@ttn-lw/components/page-title'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'

import RequireRequest from '@ttn-lw/lib/components/require-request'

import ProfileSettingsForm from '@console/containers/profile-settings-form'

import Require from '@console/lib/components/require'

import { mayViewOrEditUserSettings } from '@console/lib/feature-checks'

import { getIsConfiguration } from '@console/store/actions/identity-server'

const m = defineMessages({
  profileEdit: 'Edit profile',
})

const ProfileSettings = () => {
  useBreadcrumbs(
    'user-settings.profile',
    <Breadcrumb path={`/user-settings/profile`} content={m.profileEdit} />,
  )

  return (
    <Require featureCheck={mayViewOrEditUserSettings} otherwise={{ redirect: '/' }}>
      <RequireRequest requestAction={getIsConfiguration()}>
        <div className="container container--xl grid">
          <div className="item-12">
            <PageTitle title={m.profileEdit} />
            <ProfileSettingsForm />
          </div>
        </div>
      </RequireRequest>
    </Require>
  )
}

export default ProfileSettings
