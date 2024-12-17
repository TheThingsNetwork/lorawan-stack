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

import PageTitle from '@ttn-lw/components/page-title'
import Overlay from '@ttn-lw/components/overlay'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'

import ChangePasswordForm from '@console/containers/change-password-form'

import useRequest from '@ttn-lw/lib/hooks/use-request'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { getIsConfiguration } from '@console/store/actions/identity-server'

const ChangePassword = () => {
  const [fetching, error] = useRequest(getIsConfiguration())

  useBreadcrumbs(
    'user-settings.password',
    <Breadcrumb path={`/user-settings/password`} content={sharedMessages.changePassword} />,
  )

  if (Boolean(error)) {
    throw error
  }

  return (
    <div className="container container--xl grid">
      <div className="item-12">
        <PageTitle title={sharedMessages.changePassword} />
        <Overlay after={350} visible={fetching} loading>
          <ChangePasswordForm />
        </Overlay>
      </div>
    </div>
  )
}

export default ChangePassword
