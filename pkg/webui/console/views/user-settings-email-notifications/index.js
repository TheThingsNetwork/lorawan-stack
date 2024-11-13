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

import Message from '@ttn-lw/lib/components/message'

import EmailNotificationsForm from '@console/containers/email-notifications-form'

import Require from '@console/lib/components/require'

import { mayViewOrEditUserSettings } from '@console/lib/feature-checks'

const m = defineMessages({
  emailNotifications: 'Email notifications',
  customizeEmailNotifications:
    'Customize the notifications for which you receive emails. To see all your notifications, head to the {link}notifications panel{link}.',
})

const EmailNotificationsSettings = () => {
  useBreadcrumbs(
    'user-settings.email-notifications-settings',
    <Breadcrumb
      path={`/user-settings/email-notifications-settings`}
      content={'Email notifications settings'}
    />,
  )

  return (
    <Require featureCheck={mayViewOrEditUserSettings} otherwise={{ redirect: '/' }}>
      <div className="container container--xl grid">
        <div className="item-6 item-start-4">
          <PageTitle title={m.emailNotifications} className="mb-0" />
          <Message content={m.customizeEmailNotifications} />
          <EmailNotificationsForm />
        </div>
      </div>
    </Require>
  )
}

export default EmailNotificationsSettings
