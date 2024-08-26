// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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
import { useSelector } from 'react-redux'

import Icon, { IconKey, IconCollaborators, IconCalendarMonth } from '@ttn-lw/components/icon'
import Link from '@ttn-lw/components/link'

import Message from '@ttn-lw/lib/components/message'

import LastSeen from '@console/components/last-seen'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import { selectCollaboratorsTotalCount } from '@ttn-lw/lib/store/selectors/collaborators'

import {
  checkFromState,
  mayViewOrEditOrganizationApiKeys,
  mayViewOrEditOrganizationCollaborators,
} from '@console/lib/feature-checks'

import { selectSelectedOrganizationId } from '@console/store/selectors/organizations'
import { selectApiKeysTotalCount } from '@console/store/selectors/api-keys'

import style from './organization-header.styl'

const OrganizationHeader = ({ org }) => {
  const { name, ids, created_at } = org
  const { organization_id } = ids

  const orgId = useSelector(selectSelectedOrganizationId)
  const apiKeysTotalCount = useSelector(selectApiKeysTotalCount)
  const collaboratorsTotalCount = useSelector(state =>
    selectCollaboratorsTotalCount(state, { id: orgId }),
  )
  const mayViewCollaborators = useSelector(state =>
    checkFromState(mayViewOrEditOrganizationCollaborators, state),
  )
  const mayViewApiKeys = useSelector(state =>
    checkFromState(mayViewOrEditOrganizationApiKeys, state),
  )

  return (
    <div className={style.root} data-test-id="organization-header">
      <div>
        <h5 className={style.name}>{name || organization_id}</h5>
        <span className={style.id}>
          <Message className={style.idPrefix} content={sharedMessages.id} uppercase />
          {organization_id}
        </span>
      </div>
      <div className="d-inline-flex h-full al-center gap-cs-m flex-wrap">
        {mayViewCollaborators && (
          <Link
            to={`/organizations/${orgId}`}
            className="d-inline-flex al-center gap-cs-xxs c-text-neutral-semilight"
          >
            <Icon icon={IconCollaborators} small className="c-text-neutral-semilight" />
            <span className="fw-bold">{collaboratorsTotalCount ?? 0}</span>
            <Message
              component="span"
              content={sharedMessages.members}
              className="c-text-neutral-semilight"
            />
          </Link>
        )}
        {mayViewApiKeys && (
          <Link
            to={`/organizations/${orgId}/api-keys`}
            className="d-inline-flex al-center gap-cs-xxs c-text-neutral-semilight"
          >
            <Icon icon={IconKey} small />
            <span className="fw-bold">{apiKeysTotalCount ?? 0}</span>
            <Message component="span" content={sharedMessages.apiKeys} />
          </Link>
        )}
        <div className="d-flex al-center gap-cs-xxs sm:d-none">
          <Icon small className="c-text-neutral-semilight" icon={IconCalendarMonth} />
          <LastSeen
            statusClassName={style.createdAtStatus}
            message={sharedMessages.created}
            lastSeen={created_at}
            className="c-text-neutral-semilight"
          />
        </div>
      </div>
    </div>
  )
}

OrganizationHeader.propTypes = {
  org: PropTypes.organization.isRequired,
}

export default OrganizationHeader
