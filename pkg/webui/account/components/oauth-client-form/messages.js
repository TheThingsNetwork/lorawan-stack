// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

import { defineMessages } from 'react-intl'

export default defineMessages({
  clientIdPlaceholder: 'my-new-oauth-client',
  clientNamePlaceholder: 'My new OAuth client',
  clientDescPlaceholder: 'Description for my new OAuth client',
  clientDescDescription:
    'The description is displayed to the user when authorizing the client. Use it to explain the purpose of your client.',
  createClient: 'Create OAuth client',
  deleteTitle: 'Are you sure you want to delete this account?',
  deleteWarning:
    'This will <strong>PERMANENTLY DELETE THIS OAUTH CLIENT</strong> and <strong>LOCK THE OAUTH ID</strong>. Make sure you assign new collaborators to such entities if you plan to continue using them.',
  purgeWarning:
    'This will <strong>PERMANENTLY DELETE THIS OAUTH CLIENT</strong>. This operation cannot be undone.',
  redirectUrls: 'Redirect URLs',
  addRedirectUri: 'Add redirect URL',
  addLogoutRedirectUri: 'Add logout redirect URL',
  redirectUrlDescription:
    'The allowed redirect URIs against which authorization requests are checked',
  logoutRedirectUrls: 'Logout redirect URLs',
  logoutRedirectUrlsDescription:
    'The allowed logout redirect URIs against which client initiated logout requests are checked',
  skipAuthorization: 'Skip Authorization',
  skipAuthorizationDesc: 'If set, the authorization page will be skipped',
  endorsed: 'Endorsed',
  endorsedDesc:
    'If set, the authorization page will visually indicate endorsement to improve trust',
  grants: 'Grant types',
  grantsDesc: 'OAuth flows that can be used for the client to get a token',
  grantAuthorizationLabel: 'Authorization code',
  grantRefreshTokenLabel: 'Refresh token',
  grantPasswordLabel: 'Password',
  deleteClient: 'Delete OAuth client',
  urlsPlaceholder: 'https://example.com/oauth/callback',
  rightsWarning:
    'Note that only the minimum set of rights needed to provide the functionality of the application should be requested',
  updateWarning:
    'Note that the OAuth client will have to de re-authorized before the chosen rights are granted',
  adminOptions: 'Advanced admin options',
  grantTypeAndRights: 'Grant types and rights',
  stateDescriptionDesc:
    'You can use this field to save additional information about the state of this OAuth client, e.g. why it has been flagged',
  contactWarning:
    'Note that if no contact is provided, it will default to the first collaborator of the client.',
  adminContactDescription:
    'Administrative contact information for this client. Typically used to indicate who to contact with administrative questions about the client.',
  techContactDescription:
    'Technical contact information for this client. Typically used to indicate who to contact with technical/security questions about the client.',
  contactPlaceholder: 'Type to choose a contact',
})
