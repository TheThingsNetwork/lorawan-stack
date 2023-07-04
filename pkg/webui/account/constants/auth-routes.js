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

import Overview from '@account/views/overview'
import ProfileSettings from '@account/views/profile-settings'
import Code from '@account/views/code'
import SessionManagement from '@account/views/session-management'
import { ValidateWithAuth } from '@account/views/validate'
import OAuthClients from '@account/views/oauth-clients'
import OAuthClientAuthorizations from '@account/views/oauth-client-authorizations'

export default [
  {
    path: '/',
    Component: Overview,
  },
  {
    path: '/profile-settings',
    Component: ProfileSettings,
  },
  {
    path: '/code',
    Component: Code,
  },
  {
    path: '/session-management',
    Component: SessionManagement,
  },
  {
    path: '/validate',
    Component: ValidateWithAuth,
  },
  {
    path: '/oauth-clients/*',
    Component: OAuthClients,
  },
  {
    path: '/client-authorizations/*',
    Component: OAuthClientAuthorizations,
  },
]
