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

import React, { useCallback } from 'react'
import { defineMessages } from 'react-intl'
import { useNavigate } from 'react-router-dom'

import PageTitle from '@ttn-lw/components/page-title'
import Link from '@ttn-lw/components/link'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'

import Message from '@ttn-lw/lib/components/message'

import OrganizationAddForm from '@console/containers/organization-form/add'

import Require from '@console/lib/components/require'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { mayCreateOrganizations } from '@console/lib/feature-checks'

const m = defineMessages({
  orgDescription:
    'Organizations are used to group multiple users and assigning collective rights for them. An organization can then be set as collaborator of applications or gateways. This makes it easy to grant or revoke rights to entities for a group of users.{break} Learn more in our guide on <Link>Organization Management</Link>.',
})

const OrganizationAdd = () => {
  const navigate = useNavigate()

  const handleSuccess = useCallback(
    orgId => {
      navigate(`/organizations/${orgId}`)
    },
    [navigate],
  )

  useBreadcrumbs(
    'overview.orgs.add',
    <Breadcrumb path="/organizations/add" content={sharedMessages.createOrganization} />,
  )

  return (
    <Require featureCheck={mayCreateOrganizations} otherwise={{ redirect: '/organizations' }}>
      <div className="container container--xl grid">
        <PageTitle
          colProps={{ md: 10, lg: 9 }}
          className="mb-cs-s"
          title={sharedMessages.createOrganization}
        >
          <Message
            component="p"
            content={m.orgDescription}
            values={{
              Link: content => (
                <Link.DocLink secondary path="/the-things-stack/management/user-management/org/">
                  {content}
                </Link.DocLink>
              ),
              break: <br />,
            }}
          />
          <hr className="mb-ls-s" />
        </PageTitle>
        <div className="item-10">
          <OrganizationAddForm onSuccess={handleSuccess} />
        </div>
      </div>
    </Require>
  )
}

export default OrganizationAdd
