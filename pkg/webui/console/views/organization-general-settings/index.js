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

import React, { useCallback, useEffect, useState } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import { useNavigate } from 'react-router-dom'

import PageTitle from '@ttn-lw/components/page-title'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import Spinner from '@ttn-lw/components/spinner'

import { FullViewErrorInner } from '@ttn-lw/lib/components/full-view-error'
import Message from '@ttn-lw/lib/components/message'
import RequireRequest from '@ttn-lw/lib/components/require-request'

import OrganizationUpdateForm from '@console/containers/organization-form/update'

import Require from '@console/lib/components/require'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import { getCollaboratorsList } from '@ttn-lw/lib/store/actions/collaborators'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import {
  mayEditBasicOrganizationInformation,
  mayDeleteOrganization,
  mayViewOrEditOrganizationApiKeys,
  mayViewOrEditOrganizationCollaborators,
  checkFromState,
} from '@console/lib/feature-checks'

import { getApiKeysList } from '@console/store/actions/api-keys'
import { getIsConfiguration } from '@console/store/actions/identity-server'

import { selectSelectedOrganizationId } from '@console/store/selectors/organizations'

const GeneralSettings = () => {
  const [error, setError] = useState()
  const [fetching, setFetching] = useState(true)
  const dispatch = useDispatch()
  const navigate = useNavigate()
  const mayDeleteOrg = useSelector(state => checkFromState(mayDeleteOrganization, state))
  const mayViewApiKeys = useSelector(state =>
    checkFromState(mayViewOrEditOrganizationApiKeys, state),
  )
  const mayViewCollaborators = useSelector(state =>
    checkFromState(mayViewOrEditOrganizationCollaborators, state),
  )
  const orgId = useSelector(selectSelectedOrganizationId)

  useBreadcrumbs(
    'overview.orgs.single.general-settings',
    <Breadcrumb
      path={`/organizations/${orgId}/general-settings`}
      content={sharedMessages.generalSettings}
    />,
  )

  // Conditionally load API Keys and Collaborators to determine whether
  // deleting is possible.
  useEffect(() => {
    try {
      if (mayDeleteOrg) {
        if (mayViewApiKeys) {
          dispatch(attachPromise(getApiKeysList('organization', orgId)))
        }

        if (mayViewCollaborators) {
          dispatch(attachPromise(getCollaboratorsList('organization', orgId)))
        }
      }
    } catch (error) {
      setError(error)
    }
    setFetching(false)
  }, [dispatch, mayDeleteOrg, mayViewApiKeys, mayViewCollaborators, orgId])

  const handleDeleteSuccess = useCallback(() => navigate(`/organizations`), [navigate])

  if (fetching) {
    return (
      <Spinner inline center>
        <Message content={sharedMessages.fetching} />
      </Spinner>
    )
  }

  if (error) {
    return <FullViewErrorInner error={error} />
  }

  return (
    <Require
      featureCheck={mayEditBasicOrganizationInformation}
      otherwise={{ redirect: `/organizations/${orgId}` }}
    >
      <RequireRequest requestAction={getIsConfiguration()}>
        <div className="container container--xl grid">
          <PageTitle title={sharedMessages.generalSettings} />
          <div className="item-12">
            <OrganizationUpdateForm onDeleteSuccess={handleDeleteSuccess} />
          </div>
        </div>
      </RequireRequest>
    </Require>
  )
}

export default GeneralSettings
