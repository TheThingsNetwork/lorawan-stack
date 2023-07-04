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
import { Container, Col, Row } from 'react-grid-system'
import { useParams } from 'react-router-dom'

import { ORGANIZATION } from '@console/constants/entities'

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import PageTitle from '@ttn-lw/components/page-title'

import RequireRequest from '@ttn-lw/lib/components/require-request'

import { ApiKeyEditForm } from '@console/containers/api-key-form'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { getApiKey } from '@console/store/actions/api-keys'

const OrganizationApiKeyEditInner = () => {
  const { orgId, apiKeyId } = useParams()

  useBreadcrumbs(
    'orgs.single.api-keys.edit',
    <Breadcrumb
      path={`/organizations/${orgId}/api-keys/${apiKeyId}`}
      content={sharedMessages.edit}
    />,
  )

  return (
    <Container>
      <PageTitle title={sharedMessages.keyEdit} />
      <Row>
        <Col lg={8} md={12}>
          <ApiKeyEditForm entity={ORGANIZATION} entityId={orgId} />
        </Col>
      </Row>
    </Container>
  )
}

const OrganizationApiKeyEdit = () => {
  const { orgId, apiKeyId } = useParams()

  return (
    <RequireRequest requestAction={getApiKey('organization', orgId, apiKeyId)}>
      <OrganizationApiKeyEditInner />
    </RequireRequest>
  )
}

export default OrganizationApiKeyEdit
