// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import sharedMessages from '../../../lib/shared-messages'

@withBreadcrumb('apps.single.access.single', function (props) {
  const { match } = props
  const appId = match.params.appId
  const keyId = match.params.apiKeyId

  return (
    <Breadcrumb
      path={`/console/applications/${appId}/access/${keyId}/edit`}
      icon="general_settings"
      content={sharedMessages.edit}
    />
  )
})
export default class ApplicationAccessEdit extends React.Component {

  render () {
    return <div>edit app access</div>
  }
}
