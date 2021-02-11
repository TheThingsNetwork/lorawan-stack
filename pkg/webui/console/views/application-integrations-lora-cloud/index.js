// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
import { Container, Col, Row } from 'react-grid-system'
import { defineMessages } from 'react-intl'
import { useSelector } from 'react-redux'

import LoRaCloudImage from '@assets/misc/lora-cloud.png'

import PageTitle from '@ttn-lw/components/page-title'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'
import Link from '@ttn-lw/components/link'

import Message from '@ttn-lw/lib/components/message'
import ErrorView from '@ttn-lw/lib/components/error-view'

import LoRaCloudForm from '@console/containers/lora-cloud-form'

import Require from '@console/lib/components/require'

import SubViewError from '@console/views/sub-view-error'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { mayViewOrEditApplicationPackages } from '@console/lib/feature-checks'

import { selectSelectedApplicationId } from '@console/store/selectors/applications'

import style from './application-integrations-lora-cloud.styl'

const m = defineMessages({
  officialLoRaCloudDocumentation: 'Official LoRa Cloud documentation',
  loRaCloudInfoText: `With the LoRa Cloud Device & Application Services protocol, you can manage common device functionality at the application layer for LoRaWAN®-enabled devices. This protocol consists of a set of messages that are exchanged on a predefined device management LoRaWAN port (199 by default). The purpose of these messages is three-fold:
<ol><li>Periodically communicate info messages</li><li>Trigger client-initiated management commands</li><li>Run advanced, application-layer protocols which solve common LoRaWAN use cases</li></ol>`,
  furtherResources: 'Further resources',
  setToken: 'Set LoRa Cloud token',
})

const LoRaCloud = () => {
  const appId = useSelector(selectSelectedApplicationId)
  return (
    <Require
      featureCheck={mayViewOrEditApplicationPackages}
      otherwise={{ redirect: `/applications/${appId}` }}
    >
      <ErrorView ErrorComponent={SubViewError}>
        <Container>
          <PageTitle title="LoRa Cloud Device & Application Services" />
          <Row>
            <Col lg={8} md={12}>
              <img className={style.logo} src={LoRaCloudImage} alt="LoRa Cloud" />
              <Message
                content={m.loRaCloudInfoText}
                className={style.info}
                values={{
                  ol: msg => <ol key="list">{msg}</ol>,
                  li: msg => <li>{msg}</li>,
                }}
              />
              <div>
                <Message
                  component="h4"
                  content={m.furtherResources}
                  className={style.furtherResources}
                />
                <Link.DocLink
                  path="/integrations/application-packages/lora-cloud-device-and-application-services/"
                  secondary
                >
                  LoRa Cloud Device & Application Services
                </Link.DocLink>
                {' | '}
                <Link.Anchor
                  href="https://www.loracloud.com/documentation/device_management"
                  external
                  secondary
                >
                  <Message content={m.officialLoRaCloudDocumentation} />
                </Link.Anchor>
              </div>
              <hr className={style.hRule} />
              <Message component="h3" content={m.setToken} />
              <LoRaCloudForm />
            </Col>
          </Row>
        </Container>
      </ErrorView>
    </Require>
  )
}

export default withBreadcrumb('apps.single.integrations.lora-cloud', function (props) {
  const { appId } = props

  return (
    <Breadcrumb
      path={`/applications/${appId}/integrations/lora-cloud`}
      content={sharedMessages.loraCloud}
    />
  )
})(LoRaCloud)
