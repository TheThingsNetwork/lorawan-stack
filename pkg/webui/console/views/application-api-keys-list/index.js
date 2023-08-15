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
import { Container, Row, Col } from 'react-grid-system'
import { useParams } from 'react-router-dom'
import { createSelector } from 'reselect'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'

import ApiKeysTable from '@console/containers/api-keys-table'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { getApiKeysList } from '@console/store/actions/api-keys'

import { selectApiKeys, selectApiKeysTotalCount } from '@console/store/selectors/api-keys'

const ApplicationApiKeysList = () => {
  const { appId } = useParams()

  const getApiKeys = React.useCallback(
    filters => getApiKeysList('application', appId, filters),
    [appId],
  )

  const baseDataSelectors = createSelector(
    [selectApiKeys, selectApiKeysTotalCount],
    (keys, totalCount) => ({
      keys,
      totalCount,
    }),
  )

  const baseDataSelector = useCallback(
    state => baseDataSelectors(state, appId),
    [appId, baseDataSelectors],
  )

  return (
    <Container>
      <Row>
        <IntlHelmet title={sharedMessages.apiKeys} />
        <Col>
          <ApiKeysTable baseDataSelector={baseDataSelector} getItemsAction={getApiKeys} />
        </Col>
      </Row>
    </Container>
  )
}

export default ApplicationApiKeysList
