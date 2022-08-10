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
import { Container, Col, Row } from 'react-grid-system'
import { defineMessages } from 'react-intl'
import { useDispatch } from 'react-redux'
import { push } from 'connected-react-router'

import PageTitle from '@ttn-lw/components/page-title'
import Link from '@ttn-lw/components/link'

import Message from '@ttn-lw/lib/components/message'
import RequireRequest from '@ttn-lw/lib/components/require-request'

import GatewayOnboardingForm from '@console/containers/gateway-onboarding-form'

import Require from '@console/lib/components/require'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { mayCreateGateways } from '@console/lib/feature-checks'

import { getOrganizationsList } from '@console/store/actions/organizations'

const m = defineMessages({
  gtwOnboardingDescription:
    'Register your gateway to enable data traffic between nearby end devices and the network. {break} Learn more in our {gatewayGuideURL}.',
})

const GatewayGuideLink = (
  <Link.Anchor
    external
    secondary
    href="https://www.thethingsindustries.com/docs/gateways/adding-gateways/"
  >
    Gateway Guide
  </Link.Anchor>
)

const GatewayAdd = () => {
  const dispatch = useDispatch()
  const handleSuccess = useCallback(
    gtwId => {
      dispatch(push(`/gateways/${gtwId}`))
    },
    [dispatch],
  )

  return (
    <Require featureCheck={mayCreateGateways} otherwise={{ redirect: '/gateways' }}>
      <RequireRequest requestAction={getOrganizationsList()}>
        <Container>
          <PageTitle
            colProps={{ md: 10, lg: 9 }}
            className="mb-cs-s"
            title={sharedMessages.registerGateway}
          >
            <Message
              component="p"
              content={m.gtwOnboardingDescription}
              values={{ gatewayGuideURL: GatewayGuideLink, break: <br /> }}
            />
            <hr className="mb-ls-s" />
          </PageTitle>
          <Row>
            <Col md={10} lg={9}>
              <GatewayOnboardingForm onSuccess={handleSuccess} />
            </Col>
          </Row>
        </Container>
      </RequireRequest>
    </Require>
  )
}

export default GatewayAdd
