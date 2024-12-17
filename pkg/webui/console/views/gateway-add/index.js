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
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'

import Message from '@ttn-lw/lib/components/message'
import RequireRequest from '@ttn-lw/lib/components/require-request'

import GatewayOnboardingForm from '@console/containers/gateway-onboarding-form'

import Require from '@console/lib/components/require'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { mayCreateGateways } from '@console/lib/feature-checks'

import { getOrganizationsList } from '@console/store/actions/organizations'

const m = defineMessages({
  gtwOnboardingDescription:
    'Register your gateway to enable data traffic between nearby end devices and the network. {break} Learn more in our guide on <Link>Adding Gateways</Link>.',
})

const GatewayGuideLink = content => (
  <Link.DocLink secondary path="/gateways/adding-gateways">
    {content}
  </Link.DocLink>
)

const GatewayAdd = () => {
  const navigate = useNavigate()
  const handleSuccess = useCallback(
    (gtwId, isManaged = false) => {
      if (isManaged) {
        navigate(`/gateways/${gtwId}/managed-gateway/connection-settings?claimed=true`)
        return
      }
      navigate(`/gateways/${gtwId}`)
    },
    [navigate],
  )

  useBreadcrumbs(
    'gtws.add',
    <Breadcrumb path={`/gateways/add`} content={sharedMessages.registerGateway} />,
  )

  return (
    <Require featureCheck={mayCreateGateways} otherwise={{ redirect: '/gateways' }}>
      <RequireRequest requestAction={getOrganizationsList()}>
        <div className="container container--xl grid">
          <PageTitle className="mb-cs-s" title={sharedMessages.registerGateway}>
            <Message
              component="p"
              content={m.gtwOnboardingDescription}
              values={{ Link: GatewayGuideLink, break: <br /> }}
            />
          </PageTitle>
          <div className="item-12">
            <GatewayOnboardingForm onSuccess={handleSuccess} />
          </div>
        </div>
      </RequireRequest>
    </Require>
  )
}

export default GatewayAdd
