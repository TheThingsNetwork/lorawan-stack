// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

import Overview from '@account/views/overview'
import ProfileSettings from '@account/views/profile-settings'
import Code from '@account/views/code'
import SessionManagement from '@account/views/session-management'
import { ValidateWithAuth } from '@account/views/validate'
import ClientsList from '@account/views/oauth-clients-list'
import OAuthClientAdd from '@account/views/oauth-client-add'
import OAuthClient from '@account/views/oauth-client'

export default [
  {
    path: '/',
    exact: true,
    component: Overview,
  },
  {
    path: '/profile-settings',
    exact: true,
    component: ProfileSettings,
  },
  {
    path: '/code',
    exact: true,
    component: Code,
  },
  {
    path: '/session-management',
    exact: true,
    component: SessionManagement,
  },
  {
    path: '/validate',
    exact: true,
    component: ValidateWithAuth,
  },
  {
    path: '/oauth-clients',
    exact: true,
    component: ClientsList,
  },
  {
    path: '/oauth-clients/add',
    exact: true,
    component: OAuthClientAdd,
  },
  {
    path: '/oauth-clients/:id',
    exact: true,
    component: OAuthClient,
  },
]
