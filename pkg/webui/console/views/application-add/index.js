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

import React, { useCallback } from 'react'
import { defineMessages } from 'react-intl'
import { useNavigate } from 'react-router-dom'

import PageTitle from '@ttn-lw/components/page-title'
import Link from '@ttn-lw/components/link'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'

import RequireRequest from '@ttn-lw/lib/components/require-request'
import Message from '@ttn-lw/lib/components/message'

import ApplicationForm from '@console/containers/applications-form'

import Require from '@console/lib/components/require'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { mayCreateApplications } from '@console/lib/feature-checks'

import { getOrganizationsList } from '@console/store/actions/organizations'

const m = defineMessages({
  appDescription:
    'Within applications, you can register and manage end devices and their network data. After setting up your device fleet, use one of our many integration options to pass relevant data to your external services.{break}Learn more in our guide on <Link>Adding Applications</Link>.',
})

const ApplicationAdd = () => {
  const navigate = useNavigate()
  const handleSuccess = useCallback(
    gtwId => {
      navigate(`/applications/${gtwId}`)
    },
    [navigate],
  )

  useBreadcrumbs(
    'apps.add',
    <Breadcrumb path={`/applications/add`} content={sharedMessages.createApplication} />,
  )

  return (
    <Require featureCheck={mayCreateApplications} otherwise={{ redirect: '/applications' }}>
      <RequireRequest requestAction={getOrganizationsList()}>
        <div className="container container--xl grid">
          <PageTitle
            colProps={{ md: 10, lg: 9 }}
            className="mb-cs-s"
            title={sharedMessages.createApplication}
          >
            <Message
              component="p"
              content={m.appDescription}
              values={{
                Link: content => (
                  <Link.DocLink secondary path="/integrations/adding-applications">
                    {content}
                  </Link.DocLink>
                ),
                break: <br />,
              }}
            />
            <hr className="mb-ls-s" />
          </PageTitle>
          <div className="item-12">
            <ApplicationForm onSuccess={handleSuccess} />
          </div>
        </div>
      </RequireRequest>
    </Require>
  )
}

export default ApplicationAdd
