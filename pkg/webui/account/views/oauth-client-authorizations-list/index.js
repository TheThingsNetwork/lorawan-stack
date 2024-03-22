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

import React from 'react'

import PageTitle from '@ttn-lw/components/page-title'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'

import OAuthClientAuthorizationsTable from '@account/containers/authorizations-table'

import sharedMessages from '@ttn-lw/lib/shared-messages'

const AuthorizationsList = () => (
  <div className="container container--lg grid">
    <IntlHelmet title={sharedMessages.oauthClientAuthorizations} />
    <div className="item-12">
      <PageTitle title={sharedMessages.oauthClientAuthorizations} hideHeading />
      <OAuthClientAuthorizationsTable />
    </div>
  </div>
)

export default AuthorizationsList
