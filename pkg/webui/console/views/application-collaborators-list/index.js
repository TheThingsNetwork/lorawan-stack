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
import { useParams } from 'react-router-dom'
import { createSelector } from 'reselect'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'

import CollaboratorsTable from '@console/containers/collaborators-table'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import {
  selectCollaborators,
  selectCollaboratorsTotalCount,
} from '@ttn-lw/lib/store/selectors/collaborators'
import { getCollaboratorsList } from '@ttn-lw/lib/store/actions/collaborators'

const ApplicationCollaboratorsList = () => {
  const { appId } = useParams()

  const getItemsAction = useCallback(
    filters => getCollaboratorsList('application', appId, filters),
    [appId],
  )

  const baseDataSelectors = createSelector(
    [selectCollaborators, selectCollaboratorsTotalCount],
    (collaborators, totalCount) => ({
      collaborators,
      totalCount,
    }),
  )

  const baseDataSelector = useCallback(
    state => baseDataSelectors(state, appId),
    [baseDataSelectors, appId],
  )

  return (
    <div className="container container--xxl p-0">
      <IntlHelmet title={sharedMessages.collaborators} />
      <CollaboratorsTable baseDataSelector={baseDataSelector} getItemsAction={getItemsAction} />
    </div>
  )
}

export default ApplicationCollaboratorsList
